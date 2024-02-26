package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Blocks interface {
	GetBlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error)
	GetBlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error)
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
