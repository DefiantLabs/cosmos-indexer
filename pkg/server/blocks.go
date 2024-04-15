// Package server implements a simple web server.
package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/shopspring/decimal"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/service"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DefiantLabs/cosmos-indexer/proto"
)

type blocksServer struct {
	pb.UnimplementedBlocksServiceServer
	srv   service.Blocks
	srvTx service.Txs
}

func NewBlocksServer(srv service.Blocks, srvTx service.Txs) *blocksServer {
	return &blocksServer{srv: srv, srvTx: srvTx}
}

func (r *blocksServer) BlockInfo(ctx context.Context, in *pb.GetBlockInfoRequest) (*pb.GetBlockInfoResponse, error) {
	res, err := r.srv.BlockInfo(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockInfoResponse{}, err
	}

	return &pb.GetBlockInfoResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, Info: r.blockToProto(res)}, nil
}

func (r *blocksServer) BlockValidators(ctx context.Context, in *pb.GetBlockValidatorsRequest) (*pb.GetBlockValidatorsResponse, error) {
	res, err := r.srv.BlockValidators(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockValidatorsResponse{}, err
	}

	return &pb.GetBlockValidatorsResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, ValidatorsList: res}, nil
}

func (r *blocksServer) TxChartByDay(ctx context.Context, in *pb.TxChartByDayRequest) (*pb.TxChartByDayResponse, error) {
	res, err := r.srvTx.ChartTxByDay(ctx, in.From.AsTime(), in.To.AsTime())
	if err != nil {
		return &pb.TxChartByDayResponse{}, err
	}
	data := make([]*pb.TxByDay, 0)
	for _, tx := range res {
		data = append(data, &pb.TxByDay{
			TxNum: tx.TxNum,
			Day:   timestamppb.New(tx.Day),
		})
	}

	return &pb.TxChartByDayResponse{TxByDay: data}, nil
}

func (r *blocksServer) TxByHash(ctx context.Context, in *pb.TxByHashRequest) (*pb.TxByHashResponse, error) {
	res, err := r.srvTx.GetTxByHash(ctx, in.Hash)
	if err != nil {
		return &pb.TxByHashResponse{}, err
	}
	return &pb.TxByHashResponse{
		Tx: r.txToProto(res),
	}, nil
}

func (r *blocksServer) TotalTransactions(ctx context.Context, in *pb.TotalTransactionsRequest) (*pb.TotalTransactionsResponse, error) {
	res, err := r.srvTx.TotalTransactions(ctx, in.To.AsTime())
	if err != nil {
		return &pb.TotalTransactionsResponse{}, err
	}
	return &pb.TotalTransactionsResponse{
		Total:     fmt.Sprintf("%d", res.Total),
		Total24H:  fmt.Sprintf("%d", res.Total24H),
		Total30D:  fmt.Sprintf("%d", res.Total30D),
		Volume24H: res.Volume24H.String(),
		Volume30D: res.Volume30D.String(),
	}, nil
}

func (r *blocksServer) Transactions(ctx context.Context, in *pb.TransactionsRequest) (*pb.TransactionsResponse, error) {
	transactions, total, err := r.srvTx.Transactions(ctx, in.Limit.Offset, in.Limit.Limit)
	if err != nil {
		return &pb.TransactionsResponse{}, err
	}

	res := make([]*pb.TxByHash, 0)
	for _, tx := range transactions {
		transaction := tx
		res = append(res, r.txToProto(transaction))
	}

	return &pb.TransactionsResponse{Tx: res,
		Result: &pb.Result{Limit: in.Limit.Limit, Offset: in.Limit.Offset, All: total}}, nil
}

func (r *blocksServer) TotalBlocks(ctx context.Context, in *pb.TotalBlocksRequest) (*pb.TotalBlocksResponse, error) {
	blocks, err := r.srv.TotalBlocks(ctx, in.To.AsTime())
	if err != nil {
		return &pb.TotalBlocksResponse{}, err
	}
	return &pb.TotalBlocksResponse{
		Height:      blocks.BlockHeight,
		Count24H:    blocks.Count24H,
		Time:        blocks.BlockTime,
		TotalFee24H: blocks.TotalFee24H.String(),
	}, nil
}

func (r *blocksServer) GetBlocks(ctx context.Context, in *pb.GetBlocksRequest) (*pb.GetBlocksResponse, error) {
	blocks, all, err := r.srv.Blocks(ctx, in.Limit.Limit, in.Limit.Offset)
	if err != nil {
		return &pb.GetBlocksResponse{}, err
	}

	res := make([]*pb.Block, 0)
	for _, bl := range blocks {
		res = append(res, r.blockToProto(bl))
	}
	return &pb.GetBlocksResponse{Blocks: res, Result: &pb.Result{Limit: in.Limit.Limit, Offset: in.Limit.Offset, All: all}}, nil
}

func (r *blocksServer) blockToProto(bl *model.BlockInfo) *pb.Block {
	return &pb.Block{
		BlockHeight:       bl.BlockHeight,
		ProposedValidator: bl.ProposedValidatorAddress,
		GenerationTime:    timestamppb.New(time.Now()), // TODO
		TxHash:            bl.BlockHash,
		TotalTx:           bl.TotalTx,
		GasUsed:           bl.GasUsed.String(),
		GasWanted:         bl.GasWanted.String(),
		TotalFees:         bl.TotalFees.String(),
		BlockRewards:      decimal.NewFromInt(0).String(),
	}
}

func (r *blocksServer) BlockSignatures(ctx context.Context, in *pb.BlockSignaturesRequest) (*pb.BlockSignaturesResponse, error) {
	signs, all, err := r.srv.BlockSignatures(ctx, in.BlockHeight, in.Limit.Limit, in.Limit.Offset)
	if err != nil {
		return &pb.BlockSignaturesResponse{}, err
	}

	data := make([]*pb.SignerAddress, 0)
	for _, sign := range signs {
		data = append(data, r.blockSignToProto(sign))
	}

	return &pb.BlockSignaturesResponse{
		Signers: data,
		Result: &pb.Result{
			Limit:  in.Limit.Limit,
			Offset: in.Limit.Offset,
			All:    all,
		},
	}, nil
}

func (r *blocksServer) blockSignToProto(in *model.BlockSigners) *pb.SignerAddress {
	return &pb.SignerAddress{
		Address: in.Validator,
		Time:    timestamppb.New(in.Time),
		Rank:    in.Rank,
	}
}

func (r *blocksServer) TxsByBlock(ctx context.Context, in *pb.TxsByBlockRequest) (*pb.TxsByBlockResponse, error) {
	transactions, all, err := r.srvTx.TransactionsByBlock(ctx, in.BlockHeight, in.Limit.Limit, in.Limit.Offset)
	if err != nil {
		return &pb.TxsByBlockResponse{}, err
	}

	data := make([]*pb.TxByHash, 0)
	for _, tx := range transactions {
		transaction := tx
		data = append(data, r.txToProto(transaction))
	}

	return &pb.TxsByBlockResponse{
		Data: data,
		Result: &pb.Result{
			Limit:  in.Limit.Limit,
			Offset: in.Limit.Offset,
			All:    all,
		},
	}, nil
}

func (s *blocksServer) txToProto(tx *models.Tx) *pb.TxByHash {
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
		Fees: s.toFeesProto(tx.Fees),
		TxResponse: &pb.TxResponse{
			Height:    tx.TxResponse.Height,
			Txhash:    tx.Hash,
			Codespace: tx.TxResponse.Codespace,
			Code:      int32(tx.TxResponse.Code),
			Data:      tx.TxResponse.Data,
			Info:      tx.TxResponse.Info,
			RawLog:    base64.StdEncoding.EncodeToString(tx.TxResponse.RawLog),
			GasWanted: fmt.Sprintf("%d", tx.TxResponse.GasWanted),
			GasUsed:   fmt.Sprintf("%d", tx.TxResponse.GasUsed),
			Timestamp: tx.TxResponse.TimeStamp,
		},
		Block: s.toBlockProto(&tx.Block),
	}
}

func (s *blocksServer) toFeesProto(fees []models.Fee) []*pb.Fee {
	res := make([]*pb.Fee, 0)
	for _, fee := range fees {
		res = append(res, &pb.Fee{
			Amount: fee.Amount.String(),
			Denom:  fee.Denomination.Base,
			Payer:  fee.PayerAddress.Address,
		})
	}
	return res
}

func (s *blocksServer) toBlockProto(bl *models.Block) *pb.Block {
	return &pb.Block{
		BlockHeight:       bl.Height,
		ProposedValidator: bl.ProposerConsAddress.Address,
		TxHash:            bl.BlockHash,
		GenerationTime:    timestamppb.New(bl.TimeStamp),
	}
}

func (s *blocksServer) txTipToProto(tips []models.TipAmount) []*pb.Denom {
	denoms := make([]*pb.Denom, 0)
	for _, tip := range tips {
		denoms = append(denoms, &pb.Denom{
			Denom:  tip.Denom,
			Amount: tip.Amount.String(),
		})
	}
	return denoms
}

func (s *blocksServer) toSignerInfosProto(signs []*models.SignerInfo) []*pb.SignerInfo {
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
