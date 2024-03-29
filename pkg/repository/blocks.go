package repository

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Blocks interface {
	GetBlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error)
	GetBlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error)
	TotalBlocks(ctx context.Context, to time.Time) (*model.TotalBlocks, error)
	Blocks(ctx context.Context, limit int64, offset int64) ([]*model.BlockInfo, int64, error)
}

type blocks struct {
	db *pgxpool.Pool
}

func NewBlocks(db *pgxpool.Pool) Blocks {
	return &blocks{db: db}
}

func (r *blocks) GetBlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error) {
	query := `
				SELECT bl.id, bl.height, addr.address as proposed_validator, bl.time_stamp, bl.block_hash
				from blocks bl 
				LEFT JOIN addresses addr on bl.proposer_cons_address_id = addr.id
				where bl.chain_id=$1 and bl.height = $2
				`
	o := new(model.BlockInfo)
	var blockID int64
	err := r.db.QueryRow(ctx, query, chainID, block).Scan(
		&blockID,
		&o.BlockHeight,
		&o.ProposedValidatorAddress,
		&o.GenerationTime,
		&o.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("exec %v", err)
	}

	queryTotalFees := `select sum(amount) from fees where tx_id IN (select id from txes where block_id=$1)`
	var totalFees decimal.Decimal
	err = r.db.QueryRow(ctx, queryTotalFees, blockID).Scan(&totalFees)
	if err != nil {
		return nil, fmt.Errorf("exec total fees %v", err)
	}
	o.TotalFees = totalFees

	return o, nil
}

func (r *blocks) GetBlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error) {
	query := `
				SELECT addr.address
				FROM blocks bl
				INNER JOIN txes tx on bl.id = tx.block_id
				INNER JOIN tx_signer_addresses signs on tx.id = signs.tx_id
				INNER JOIN addresses addr on signs.address_id = addr.id
				where bl.height = $1 and bl.chain_id = $2
				`
	rows, err := r.db.Query(ctx, query, block, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make([]string, 0)

	for rows.Next() {
		var in string
		errScan := rows.Scan(&in)
		if errScan != nil {
			return nil, fmt.Errorf("repository.GetBlockValidators, Scan: %v", errScan)
		}
		data = append(data, in)
	}

	return data, nil
}

func (r *blocks) TotalBlocks(ctx context.Context, to time.Time) (*model.TotalBlocks, error) {
	query := `select blocks.height from blocks order by blocks.height desc limit 1`
	row := r.db.QueryRow(ctx, query)
	var blockHeight int64
	if err := row.Scan(&blockHeight); err != nil {
		return nil, err
	}

	from := to.Add(-24 * time.Hour)
	query = `select count(*) from blocks where blocks.time_stamp between $1 AND $2`
	row = r.db.QueryRow(ctx, query, from, to)
	var count24H int64
	if err := row.Scan(&count24H); err != nil {
		return nil, err
	}

	blockTime := 0 // TODO understand how to calculate

	query = `select COALESCE(sum(fees.amount),0)
			from fees where fees.tx_id IN (
			select id from txes where block_id IN 
			(select blocks.id from blocks where blocks.time_stamp between $1 AND $2))`
	row = r.db.QueryRow(ctx, query, from, to)
	feeSum := int64(0)
	if err := row.Scan(&feeSum); err != nil {
		log.Err(err).Msgf("row.Scan(&feeSum)")
		return nil, err
	}

	return &model.TotalBlocks{
		BlockHeight: blockHeight,
		Count24H:    count24H,
		BlockTime:   int64(blockTime),
		TotalFee24H: decimal.NewFromInt(feeSum),
	}, nil
}

func (r *blocks) Blocks(ctx context.Context, limit int64, offset int64) ([]*model.BlockInfo, int64, error) {
	query := `select blocks.id, blocks.height, blocks.block_hash, addresses.address as proposer, count(txes), blocks.time_stamp from blocks
		left join addresses on blocks.proposer_cons_address_id = addresses.id
		left join txes on blocks.id = txes.block_id
		group by blocks.id, blocks.height, blocks.block_hash, addresses.address, blocks.time_stamp
		order by blocks.height desc
		limit $1 offset $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	data := make([]*model.BlockInfo, 0)
	for rows.Next() {
		var in model.BlockInfo
		blockID := 0
		errScan := rows.Scan(&blockID, &in.BlockHeight, &in.BlockHeight,
			&in.BlockHash, &in.ProposedValidatorAddress, &in.TotalTx, &in.TimeElapsed)
		if errScan != nil {
			return nil, 0, fmt.Errorf("repository.Blocks, Scan: %v", errScan)
		}

		queryFees := `select blocks.height, sum(COALESCE(fees.amount,0)) from blocks
                 left join txes on blocks.id = txes.block_id
                 left join fees on txes.id = fees.tx_id
                 where blocks.height = $1
				 group by blocks.height`
		rowFees := r.db.QueryRow(ctx, queryFees, in.BlockHeight)
		if err = rowFees.Scan(&in.BlockHeight, &in.TotalFees); err != nil {
			return nil, 0, fmt.Errorf("rowFees.Scan, Scan: %v", errScan)
		}

		queryTxs := `select count(*) from txes where txes.block_id = $1`
		rowQueryTxs := r.db.QueryRow(ctx, queryTxs, blockID)
		if err = rowQueryTxs.Scan(&in.TotalTx); err != nil {
			return nil, 0, fmt.Errorf("rowQueryTxs.Scan, Scan: %v", errScan)
		}

		queryGas := `select blocks.height, sum(COALESCE(tx_responses.gas_wanted,0)), sum(COALESCE(tx_responses.gas_used,0)) from blocks
						left join txes on blocks.id = txes.block_id
						left join tx_responses on txes.tx_response_id = tx_responses.id
						where blocks.height = $1
						group by blocks.height`
		rowQueryGas := r.db.QueryRow(ctx, queryGas, in.BlockHeight)
		if err = rowQueryGas.Scan(&in.BlockHeight, &in.GasUsed, &in.GasWanted); err != nil {
			return nil, 0, fmt.Errorf("rowQueryGas.Scan, Scan: %v", errScan)
		}
		data = append(data, &in)
	}

	query = `select count(*) from blocks`
	row := r.db.QueryRow(ctx, query)
	var all int64
	if err = row.Scan(&all); err != nil {
		return nil, 0, err
	}

	return data, all, nil
}
