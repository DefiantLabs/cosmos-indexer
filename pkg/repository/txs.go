package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shopspring/decimal"

	"github.com/DefiantLabs/cosmos-indexer/db/models"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	GetTxByHash(ctx context.Context, txHash string) (*models.Tx, error)
	TransactionsPerPeriod(ctx context.Context, to time.Time) (int64, int64, int64, error)
	VolumePerPeriod(ctx context.Context, to time.Time) (decimal.Decimal, decimal.Decimal, error)
	Transactions(ctx context.Context, limit int64, offset int64) ([]*models.Tx, int64, error)
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

func (r *txs) GetTxByHash(ctx context.Context, txHash string) (*models.Tx, error) {
	query := `
	   select 
	       txes.id as tx_id,
	       txes.signatures as signatures,
		   hash,
		   txes.code as tx_code,
		   block_id,
		   timestamp,
		   memo,
		   timeout_height,extension_options,
		   non_critical_extension_options,auth_info_id,tx_response_id,
		   auf.gas_limit, auf.payer, auf.granter, tip.tipper,
		   resp.code as tx_resp_code, resp.gas_used as tx_res_gas_used,
		   resp.gas_wanted as tx_res_gas_wanted, resp.raw_log
		from txes
			 left join tx_auth_info au on auth_info_id = au.id
			 left join tx_auth_info_fee auf on au.fee_id = auf.id
			 left join tx_tip tip on au.tip_id = tip.id
			 left join tx_responses resp on tx_response_id = resp.id
		   where hash=$1`

	var tx models.Tx
	var authInfo models.AuthInfo
	var authInfoFee models.AuthInfoFee
	var authInfoTip models.Tip
	var txResponse models.TxResponse

	row := r.db.QueryRow(ctx, query, txHash)
	if err := row.Scan(&tx.ID, &tx.Signatures, &tx.Hash, &tx.Code,
		&tx.BlockID, &tx.Timestamp, &tx.Memo, &tx.TimeoutHeight,
		&tx.ExtensionOptions, &tx.NonCriticalExtensionOptions,
		&tx.AuthInfoID, &tx.TxResponseID,
		&authInfoFee.GasLimit, &authInfoFee.Payer,
		&authInfoFee.Granter, &authInfoTip.Tipper,
		&txResponse.Code, &txResponse.GasUsed, &txResponse.GasWanted, &txResponse.RawLog); err != nil {
		return nil, err
	}

	querySignerInfos := `
						select 
						    txi.signer_info_id, 
						    txnf.address_id, 
						    txnf.mode_info, 
						    txnf.sequence, 
						    addr.address
						from tx_signer_infos txi 
						left join tx_signer_info txnf on txi.signer_info_id = txnf.id
						left join addresses addr on txnf.address_id = addr.id
                      	where txi.auth_info_id = $1`
	signerInfos := make([]*models.SignerInfo, 0)
	rows, err := r.db.Query(ctx, querySignerInfos, authInfo.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var in models.SignerInfo
		errScan := rows.Scan(&in.ID, &in.Address.ID, &in.ModeInfo, &in.Sequence, &in.Address.Address)
		if errScan != nil {
			return nil, fmt.Errorf("repository.querySignerInfos, Scan: %v", errScan)
		}
		signerInfos = append(signerInfos, &in)
	}

	authInfo.SignerInfos = signerInfos
	authInfo.Fee = authInfoFee
	authInfo.Tip = authInfoTip
	tx.AuthInfo = authInfo
	tx.TxResponse = txResponse

	return &tx, nil
}

func (r *txs) Transactions(ctx context.Context, limit int64, offset int64) ([]*models.Tx, int64, error) {
	query := `
	   select 
	       txes.id as tx_id,
	       txes.signatures as signatures,
		   hash,
		   txes.code as tx_code,
		   block_id,
		   timestamp,
		   memo,
		   timeout_height,
		   extension_options,
		   non_critical_extension_options,
		   auth_info_id,
		   tx_response_id,
		   auf.gas_limit, 
		   auf.payer, 
		   auf.granter, 
		   tip.tipper,
		   resp.code as tx_resp_code, 
		   resp.gas_used as tx_res_gas_used,
		   resp.gas_wanted as tx_res_gas_wanted, 
		   resp.raw_log, resp.time_stamp, resp.codespace, resp.data, resp.info
		from txes
			 left join tx_auth_info au on auth_info_id = au.id
			 left join tx_auth_info_fee auf on au.fee_id = auf.id
			 left join tx_tip tip on au.tip_id = tip.id
			 left join tx_responses resp on tx_response_id = resp.id
		   ORDER BY txes.timestamp desc
		   LIMIT $1 OFFSET $2`

	result := make([]*models.Tx, 0)

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		log.Err(err).Msgf("Query error")
		return nil, 0, err
	}

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

		queryBlockInfo := `
						select 
						    blocks.time_stamp, blocks.height, blocks.chain_id, addresses.address, blocks.block_hash
							from blocks 
							left join addresses on blocks.proposer_cons_address_id = addresses.id
							where blocks.id = $1
						`
		var block models.Block
		var address models.Address
		rowBlock := r.db.QueryRow(ctx, queryBlockInfo, tx.BlockID)
		if err = rowBlock.Scan(&block.TimeStamp, &block.Height, &block.ChainID, &address.Address, &block.BlockHash); err != nil {
			log.Err(err).Msgf("rowBlock.Scan error")
			return nil, 0, err
		}
		block.ProposerConsAddress = address
		tx.Block = block

		querySignerInfos := `
						select 
						    txi.signer_info_id, 
						    txnf.address_id, 
						    txnf.mode_info, 
						    txnf.sequence, 
						    addr.address
						from tx_signer_infos txi 
						left join tx_signer_info txnf on txi.signer_info_id = txnf.id
						left join addresses addr on txnf.address_id = addr.id
                      	where txi.auth_info_id = $1`
		signerInfos := make([]*models.SignerInfo, 0)
		rowsSigners, err := r.db.Query(ctx, querySignerInfos, tx.AuthInfoID)
		if err != nil {
			log.Err(err).Msgf("querySignerInfos error")
			return nil, 0, err
		}
		for rowsSigners.Next() {
			var in models.SignerInfo
			var addr models.Address

			errScan := rowsSigners.Scan(&in.ID, &addr.ID, &in.ModeInfo, &in.Sequence, &addr.Address)
			if errScan != nil {
				log.Err(err).Msgf("rowsSigners.Scan error")
				return nil, 0, fmt.Errorf("repository.querySignerInfos, Scan: %v", errScan)
			}
			in.Address = &addr
			signerInfos = append(signerInfos, &in)
		}

		authInfo.SignerInfos = signerInfos
		authInfo.Fee = authInfoFee
		authInfo.Tip = authInfoTip
		tx.AuthInfo = authInfo
		tx.TxResponse = txResponse

		result = append(result, &tx)
	}

	queryAll := `select count(*) from txes`
	row := r.db.QueryRow(ctx, queryAll)
	var allTx int64
	if err = row.Scan(&allTx); err != nil {
		log.Err(err).Msgf("queryAll error")
		return nil, 0, err
	}

	return result, allTx, nil
}
