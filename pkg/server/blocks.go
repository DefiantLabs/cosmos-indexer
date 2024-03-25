// Package server implements a simple web server.
package server

import (
	"context"
	"fmt"

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

	return &pb.GetBlockInfoResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, Info: r.toProto(res)}, nil
}

func (r *blocksServer) BlockValidators(ctx context.Context, in *pb.GetBlockValidatorsRequest) (*pb.GetBlockValidatorsResponse, error) {
	res, err := r.srv.BlockValidators(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockValidatorsResponse{}, err
	}

	return &pb.GetBlockValidatorsResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, ValidatorsList: res}, nil
}

func (r *blocksServer) toProto(in *model.BlockInfo) *pb.Block {
	return &pb.Block{
		BlockHeight:       in.BlockHeight,
		ProposedValidator: in.ProposedValidatorAddress,
		GenerationTime:    timestamppb.New(in.GenerationTime),
		TotalFees:         in.TotalFees.String(),
		TxHash:            in.BlockHash,
	}
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
	return res, nil
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
	txs, total, err := r.srvTx.Transactions(ctx, in.Limit.Offset, in.Limit.Limit)
	if err != nil {
		return &pb.TransactionsResponse{}, err
	}
	return &pb.TransactionsResponse{Tx: txs, Result: &pb.Result{Limit: in.Limit.Limit, Offset: in.Limit.Offset, All: total}}, nil
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
