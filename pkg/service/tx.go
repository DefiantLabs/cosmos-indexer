package service

import (
	"context"
	"fmt"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
	pb "github.com/DefiantLabs/cosmos-indexer/proto"
	"strings"
	"time"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error)
}

type txs struct {
	txRepo repository.Txs
}

func NewTxs(txRepo repository.Txs) *txs {
	return &txs{txRepo: txRepo}
}

func (s *txs) ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error) {
	return s.txRepo.ChartTxByDay(ctx, from, to)
}

func (s *txs) GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error) {
	tx, err := s.txRepo.GetTxByHash(ctx, strings.ToUpper(txHash))
	if err != nil {
		return &pb.TxByHashResponse{}, err
	}

	return s.txToProto(tx), nil
}

func (s *txs) txToProto(tx *models.Tx) *pb.TxByHashResponse {
	return &pb.TxByHashResponse{
		Tx: &pb.TxByHash{
			Memo:                        tx.Memo,
			TimeoutHeight:               fmt.Sprintf("%d", tx.TimeoutHeight),
			ExtensionOptions:            tx.ExtensionOptions,
			NonCriticalExtensionOptions: tx.NonCriticalExtensionOptions,
			AuthInfo: &pb.TxAuthInfo{
				PublicKey:  []string{}, // TODO
				Signatures: []string{}, // TODO
				Fee: &pb.TxFee{
					Granter:  tx.AuthInfo.Fee.Granter,
					Payer:    tx.AuthInfo.Fee.Payer,
					GasLimit: fmt.Sprintf("%d", tx.AuthInfo.Fee.GasLimit),
					Amount:   nil, // TODO
				},
				Tip: &pb.TxTip{
					Tipper: tx.AuthInfo.Tip.Tipper,
					Amount: s.txTipToProto(tx.AuthInfo.Tip.Amount),
				},
			},
			TxResponse: &pb.TxResponse{
				Height:    tx.TxResponse.Height,
				Txhash:    tx.Hash,
				Codespace: "", // TODO
				Code:      int32(tx.TxResponse.Code),
				Data:      "", // TODO
				RawLog:    tx.TxResponse.RawLog,
				Info:      "",  // TODO
				Logs:      nil, // TODO
				GasWanted: fmt.Sprintf("%d", tx.TxResponse.GasWanted),
				GasUsed:   fmt.Sprintf("%d", tx.TxResponse.GasUsed),
				Timestamp: tx.TxResponse.TimeStamp,
			},
		},
	}
}

func (s *txs) txTipToProto(tips []models.TipAmount) []*pb.Denom {
	denoms := make([]*pb.Denom, 0)
	for _, tip := range tips {
		denoms = append(denoms, &pb.Denom{
			Denom:  tip.Denom,
			Amount: tip.Amount.String(),
		})
	}
	return denoms
}
