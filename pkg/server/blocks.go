// Package server implements a simple web server.
package server

import (
	"context"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/service"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DefiantLabs/cosmos-indexer/proto"
)

type BlocksServer struct {
	pb.UnimplementedBlocksServiceServer
	srv   service.Blocks
	srvTx service.Txs
}

func NewBlocksServer(srv service.Blocks, srvTx service.Txs) *BlocksServer {
	return &BlocksServer{srv: srv, srvTx: srvTx}
}

func (r *BlocksServer) BlockInfo(ctx context.Context, in *pb.GetBlockInfoRequest) (*pb.GetBlockInfoResponse, error) {
	res, err := r.srv.BlockInfo(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockInfoResponse{}, err
	}

	return &pb.GetBlockInfoResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, Info: r.toProto(res)}, nil
}

func (r *BlocksServer) BlockValidators(ctx context.Context, in *pb.GetBlockValidatorsRequest) (*pb.GetBlockValidatorsResponse, error) {
	res, err := r.srv.BlockValidators(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockValidatorsResponse{}, err
	}

	return &pb.GetBlockValidatorsResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, ValidatorsList: res}, nil
}

func (r *BlocksServer) toProto(in *model.BlockInfo) *pb.Block {
	return &pb.Block{
		BlockHeight:       in.BlockHeight,
		ProposedValidator: in.ProposedValidatorAddress,
		GenerationTime:    timestamppb.New(in.GenerationTime),
		TotalFees:         in.TotalFees.String(),
		TxHash:            in.BlockHash,
	}
}

func (r *BlocksServer) TxChartByDay(ctx context.Context, in *pb.TxChartByDayRequest) (*pb.TxChartByDayResponse, error) {
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

func (r *BlocksServer) TxByHash(ctx context.Context, in *pb.TxByHashRequest) (*pb.TxByHashResponse, error) {
	_, err := r.srvTx.GetTxByHash(ctx, in.Hash)
	if err != nil {
		return &pb.TxByHashResponse{}, err
	}
	return nil, nil
}
