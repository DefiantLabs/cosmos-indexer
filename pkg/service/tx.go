package service

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
	pb "github.com/DefiantLabs/cosmos-indexer/proto"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error) // TODO re-implement without pb
	TotalTransactions(ctx context.Context, to time.Time) (*model.TotalTransactions, error)
	Transactions(ctx context.Context, offset int64, limit int64) ([]*pb.TxByHash, int64, error) // TODO re-implement without pb
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

func (s *txs) Transactions(ctx context.Context, offset int64, limit int64) ([]*pb.TxByHash, int64, error) {
	transactions, all, err := s.txRepo.Transactions(ctx, limit, offset)
	log.Info().Msgf("transactions len %d", len(transactions))
	if err != nil {
		return nil, 0, err
	}
	res := make([]*pb.TxByHash, 0)
	for _, tx := range transactions {
		transaction := tx
		res = append(res, s.txToProto(transaction))
	}
	return res, all, nil
}

func (s *txs) TotalTransactions(ctx context.Context, to time.Time) (*model.TotalTransactions, error) {
	var res model.TotalTransactions
	var err error
	res.Total, res.Total24H, res.Total30D, err = s.txRepo.TransactionsPerPeriod(ctx, to)
	if err != nil {
		return nil, err
	}

	res.Volume24H, res.Volume30D, err = s.txRepo.VolumePerPeriod(ctx, to)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *txs) GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error) {
	tx, err := s.txRepo.GetTxByHash(ctx, strings.ToUpper(txHash))
	if err != nil {
		return &pb.TxByHashResponse{}, err
	}
	return &pb.TxByHashResponse{Tx: s.txToProto(tx)}, nil
}

func (s *txs) txToProto(tx *models.Tx) *pb.TxByHash {
	return &pb.TxByHash{
		Memo:                        tx.Memo,
		TimeoutHeight:               fmt.Sprintf("%d", tx.TimeoutHeight),
		ExtensionOptions:            tx.ExtensionOptions,
		NonCriticalExtensionOptions: tx.NonCriticalExtensionOptions,
		AuthInfo: &pb.TxAuthInfo{
			PublicKey:  []string{}, // TODO
			Signatures: tx.Signatures,
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
			SignerInfos: s.toSignerInfosProto(tx.AuthInfo.SignerInfos),
		},
		TxResponse: &pb.TxResponse{
			Height:    tx.TxResponse.Height,
			Txhash:    tx.Hash,
			Codespace: tx.TxResponse.Codespace,
			Code:      int32(tx.TxResponse.Code),
			Data:      tx.TxResponse.Data,
			Info:      tx.TxResponse.Info,
			RawLog:    tx.TxResponse.RawLog,
			GasWanted: fmt.Sprintf("%d", tx.TxResponse.GasWanted),
			GasUsed:   fmt.Sprintf("%d", tx.TxResponse.GasUsed),
			Timestamp: tx.TxResponse.TimeStamp,
		},
		Block: s.toBlockProto(&tx.Block),
	}
}

func (s *txs) toBlockProto(bl *models.Block) *pb.Block {
	return &pb.Block{
		BlockHeight:       bl.Height,
		ProposedValidator: bl.ProposerConsAddress.Address,
		TxHash:            bl.BlockHash,
		GenerationTime:    timestamppb.New(bl.TimeStamp),
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

func (s *txs) toSignerInfosProto(signs []*models.SignerInfo) []*pb.SignerInfo {
	res := make([]*pb.SignerInfo, 0)
	for _, sign := range signs {
		res = append(res, &pb.SignerInfo{
			Address:  sign.Address.Address,
			ModeInfo: sign.ModeInfo,
			Sequence: int64(sign.Sequence),
		})
	}
	return res
}
