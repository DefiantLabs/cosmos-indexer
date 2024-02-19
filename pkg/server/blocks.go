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
	srv service.Blocks
}

func NewBlocksServer(srv service.Blocks) *BlocksServer {
	return &BlocksServer{srv: srv}
}

func (r *BlocksServer) BlockInfo(ctx context.Context, in *pb.GetBlockInfoRequest) (*pb.GetBlockInfoResponse, error) {
	res, err := r.srv.BlockInfo(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockInfoResponse{}, err
	}

	return &pb.GetBlockInfoResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId, Info: r.toProto(res)}, nil
}

func (r *BlocksServer) BlockValidators(ctx context.Context, in *pb.GetBlockValidatorsRequest) (*pb.GetBlockValidatorsResponse, error) {
	_, err := r.srv.BlockInfo(ctx, in.BlockNumber, in.ChainId)
	if err != nil {
		return &pb.GetBlockValidatorsResponse{}, err
	}

	return &pb.GetBlockValidatorsResponse{BlockNumber: in.BlockNumber, ChainId: in.ChainId}, nil
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
