package repository

import (
	"context"
	"fmt"
	pb "github.com/DefiantLabs/cosmos-indexer/proto"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
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

func (r *txs) GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error) {
	query := `select id, hash, code, block_id, timestamp from txes where hash=$1`
	_ = r.db.QueryRow(ctx, query, txHash)

	return nil, nil
}
