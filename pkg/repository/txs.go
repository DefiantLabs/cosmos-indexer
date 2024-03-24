package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/db/models"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	GetTxByHash(ctx context.Context, txHash string) (*models.Tx, error)
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
