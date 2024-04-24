package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rs/zerolog/log"

	"github.com/shopspring/decimal"

	"github.com/DefiantLabs/cosmos-indexer/db/models"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	TransactionsPerPeriod(ctx context.Context, to time.Time) (int64, int64, int64, error)
	VolumePerPeriod(ctx context.Context, to time.Time) (decimal.Decimal, decimal.Decimal, error)
	Transactions(ctx context.Context, limit int64, offset int64, filter *TxsFilter) ([]*models.Tx, int64, error)
	TransactionRawLog(ctx context.Context, hash string) ([]byte, error)
	TransactionSigners(ctx context.Context, hash string) ([]*models.SignerInfo, error)
}

type TxsFilter struct {
	TxHash        *string
	TxBlockHeight *int64
}

type txs struct {
	db *pgxpool.Pool
}

func NewTxs(db *pgxpool.Pool) Txs {
	return &txs{db: db}
}

func (r *txs) ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error) {
	query := `
				select count(txes.hash),  date_trunc('day', txes.timestamp) from txes
				where txes.timestamp >= $1 and txes.timestamp <= $2
				group by date_trunc('day', txes.timestamp)
				`
	data := make([]*model.TxsByDay, 0)
	rows, err := r.db.Query(ctx, query, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var in model.TxsByDay
		errScan := rows.Scan(&in.TxNum, &in.Day)
		if errScan != nil {
			return nil, fmt.Errorf("repository.ChartTxByDay, Scan: %v", errScan)
		}
		data = append(data, &in)
	}

	return data, nil
}

func (r *txs) TransactionsPerPeriod(ctx context.Context, to time.Time) (int64, int64, int64, error) {
	query := `select count(*) from txes`
	row := r.db.QueryRow(ctx, query)
	var allTx int64
	if err := row.Scan(&allTx); err != nil {
		return 0, 0, 0, err
	}

	from := to.UTC().Add(-24 * time.Hour)
	query = `select count(*) from txes where txes.timestamp between $1 AND $2`
	row = r.db.QueryRow(ctx, query, from.UTC(), to.UTC())
	var all24H int64
	if err := row.Scan(&all24H); err != nil {
		return 0, 0, 0, err
	}

	from = to.UTC().Add(-720 * time.Hour)
	query = `select count(*) from txes where txes.timestamp between $1 AND $2`
	row = r.db.QueryRow(ctx, query, from.UTC(), to.UTC())
	var all30D int64
	if err := row.Scan(&all30D); err != nil {
		return 0, 0, 0, err
	}

	return allTx, all24H, all30D, nil
}

func (r *txs) VolumePerPeriod(ctx context.Context, to time.Time) (decimal.Decimal, decimal.Decimal, error) {
	// TODO understand in what denom to return
	return decimal.NewFromInt(0), decimal.NewFromInt(0), nil
}

func (r *txs) TransactionSigners(ctx context.Context, txHash string) ([]*models.SignerInfo, error) {
	querySignerInfos := `
						select
							txi.signer_info_id,
							txnf.address_id,
							txnf.mode_info,
							txnf.sequence,
							addr.address
						from txes
							inner join tx_auth_info tai on txes.auth_info_id = txes.id
							inner join tx_signer_infos txi on tai.id = txi.auth_info_id
							inner join tx_signer_info txnf on txi.signer_info_id = txnf.id
							inner join addresses addr on txnf.address_id = addr.id
						where txes.hash = $1`
	signerInfos := make([]*models.SignerInfo, 0)
	rowsSigners, err := r.db.Query(ctx, querySignerInfos, txHash)
	if err != nil {
		log.Err(err).Msgf("querySignerInfos error")
		return nil, err
	}
	for rowsSigners.Next() {
		var in models.SignerInfo
		var addr models.Address

		errScan := rowsSigners.Scan(&in.ID, &addr.ID, &in.ModeInfo, &in.Sequence, &addr.Address)
		if errScan != nil {
			log.Err(err).Msgf("rowsSigners.Scan error")
			return nil, fmt.Errorf("repository.querySignerInfos, Scan: %v", errScan)
		}
		in.Address = &addr
		signerInfos = append(signerInfos, &in)
	}

	return signerInfos, nil
}

func (r *txs) TransactionRawLog(ctx context.Context, hash string) ([]byte, error) {
	query := `
	   select
		   resp.raw_log
		from txes
			 inner join tx_responses resp on txes.tx_response_id = resp.id
		where txes.hash = $1`

	row := r.db.QueryRow(ctx, query, hash)

	var rawLog []byte
	if err := row.Scan(&rawLog); err != nil {
		log.Err(err).Msgf("repository.TransactionRawLog")
		return nil, fmt.Errorf("not found")
	}

	return rawLog, nil
}

func (r *txs) Transactions(ctx context.Context, limit int64, offset int64, filter *TxsFilter) ([]*models.Tx, int64, error) {
	query := `
	   select 
	       txes.id as tx_id,
	       txes.signatures as signatures,
		   txes.hash,
		   txes.code as tx_code,
		   txes.block_id,
		   txes.timestamp,
		   txes.memo,
		   txes.timeout_height,
		   txes.extension_options,
		   txes.non_critical_extension_options,
		   txes.auth_info_id,
		   txes.tx_response_id,
		   auf.gas_limit, 
		   auf.payer, 
		   auf.granter, 
		   tip.tipper,
		   resp.code as tx_resp_code, 
		   resp.gas_used as tx_res_gas_used,
		   resp.gas_wanted as tx_res_gas_wanted, 
		   NULL as raw_log, 
		   resp.time_stamp, 
		   resp.codespace, 
		   resp.data, 
		   resp.info
		from txes
			 left join tx_auth_info au on auth_info_id = au.id
			 left join tx_auth_info_fee auf on au.fee_id = auf.id
			 left join tx_tip tip on au.tip_id = tip.id
			 left join tx_responses resp on tx_response_id = resp.id
			 left join blocks on txes.block_id = blocks.id`

	var rows pgx.Rows
	var err error
	if filter != nil { // TODO make it more flexible
		if filter.TxBlockHeight != nil {
			query += ` WHERE blocks.height = $1`
			query += ` ORDER BY txes.timestamp desc LIMIT $2 OFFSET $3`
			rows, err = r.db.Query(ctx, query, *filter.TxBlockHeight, limit, offset)
		} else if filter.TxHash != nil && len(*filter.TxHash) > 0 {
			query += ` WHERE hash = $1`
			query += ` ORDER BY txes.timestamp desc LIMIT $2 OFFSET $3`
			rows, err = r.db.Query(ctx, query, *filter.TxHash, limit, offset)
		}
	} else {
		query += ` ORDER BY txes.timestamp desc LIMIT $1 OFFSET $2`
		rows, err = r.db.Query(ctx, query, limit, offset)
	}

	if err != nil {
		log.Err(err).Msgf("Query error")
		return nil, 0, err
	}

	result := make([]*models.Tx, 0)
	for rows.Next() {
		var tx models.Tx
		var authInfo models.AuthInfo
		var authInfoFee models.AuthInfoFee
		var authInfoTip models.Tip
		var txResponse models.TxResponse
		signatures := make([][]byte, 0)
		extensionsOptions := make([]string, 0)
		nonCriticalExtensionOptions := make([]string, 0)

		if err := rows.Scan(&tx.ID, &signatures, &tx.Hash, &tx.Code,
			&tx.BlockID, &tx.Timestamp, &tx.Memo, &tx.TimeoutHeight,
			&extensionsOptions, &nonCriticalExtensionOptions,
			&tx.AuthInfoID, &tx.TxResponseID,
			&authInfoFee.GasLimit, &authInfoFee.Payer,
			&authInfoFee.Granter, &authInfoTip.Tipper,
			&txResponse.Code, &txResponse.GasUsed,
			&txResponse.GasWanted, &txResponse.RawLog,
			&txResponse.TimeStamp, &txResponse.Codespace,
			&txResponse.Data, &txResponse.Info); err != nil {
			log.Err(err).Msgf("rows.Scan error")
			return nil, 0, err
		}
		tx.Signatures = signatures
		tx.ExtensionOptions = extensionsOptions
		tx.NonCriticalExtensionOptions = nonCriticalExtensionOptions

		var block *models.Block
		if block, err = r.blockInfo(ctx, tx.BlockID); err != nil {
			log.Err(err).Msgf("error in blockInfo")
		}
		if block != nil {
			tx.Block = *block
		}

		var fees []models.Fee
		if fees, err = r.feesByTransaction(ctx, tx.ID); err != nil {
			log.Err(err).Msgf("error in feesByTransaction")
		}
		tx.Fees = fees

		authInfo.Fee = authInfoFee
		authInfo.Tip = authInfoTip
		tx.AuthInfo = authInfo
		tx.TxResponse = txResponse

		result = append(result, &tx)
	}

	blockID := -1
	if filter != nil {
		queryBlock := `select id from blocks where blocks.height = $1`
		row := r.db.QueryRow(ctx, queryBlock, *filter.TxBlockHeight)
		if err = row.Scan(&blockID); err != nil {
			log.Err(err).Msgf("queryBlock error")
		}
	}

	var row pgx.Row
	if blockID >= 0 {
		queryAll := `select count(*) from txes where txes.block_id = $1`
		row = r.db.QueryRow(ctx, queryAll, blockID)
	} else {
		queryAll := `select count(*) from txes`
		row = r.db.QueryRow(ctx, queryAll)
	}

	var allTx int64
	if err = row.Scan(&allTx); err != nil {
		log.Err(err).Msgf("queryAll error")
		return nil, 0, err
	}

	return result, allTx, nil
}

func (r *txs) feesByTransaction(ctx context.Context, txID uint) ([]models.Fee, error) {
	feesQ := `select fs.amount, dm.base, a.address from fees fs 
    				left join public.addresses a on fs.payer_address_id = a.id
    				left join denoms dm on fs.denomination_id = dm.id
    				where fs.tx_id = $1`
	rowsFees, err := r.db.Query(ctx, feesQ, txID)
	if err != nil {
		log.Err(err).Msgf("feesRes error")
		return nil, err
	}
	feesRes := make([]models.Fee, 0)
	for rowsFees.Next() {
		var in models.Fee
		var denom models.Denom
		var address models.Address

		errScan := rowsFees.Scan(&in.Amount, &denom.Base, &address.Address)
		if errScan != nil {
			log.Err(err).Msgf("rowsFees.Scan error")
			return nil, fmt.Errorf("repository.feesByTransaction, Scan: %v", errScan)
		}
		in.Denomination = denom
		in.PayerAddress = address
		feesRes = append(feesRes, in)
	}
	return feesRes, nil
}

func (r *txs) blockInfo(ctx context.Context, blockID uint) (*models.Block, error) {
	queryBlockInfo := `
						select 
						    blocks.time_stamp, 
						    blocks.height, 
						    blocks.chain_id, 
						    addresses.address, 
						    blocks.block_hash
							from blocks 
							left join addresses on blocks.proposer_cons_address_id = addresses.id
							where blocks.id = $1
						`
	var block models.Block
	var address models.Address
	rowBlock := r.db.QueryRow(ctx, queryBlockInfo, blockID)
	if err := rowBlock.Scan(&block.TimeStamp, &block.Height, &block.ChainID, &address.Address, &block.BlockHash); err != nil {
		log.Err(err).Msgf("rowBlock.Scan error")
		return nil, err
	}
	block.ProposerConsAddress = address
	return &block, nil
}
