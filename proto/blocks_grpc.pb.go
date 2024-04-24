// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: blocks.proto

package blocks

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	BlocksService_BlockInfo_FullMethodName          = "/blocks.BlocksService/BlockInfo"
	BlocksService_BlockValidators_FullMethodName    = "/blocks.BlocksService/BlockValidators"
	BlocksService_TxChartByDay_FullMethodName       = "/blocks.BlocksService/TxChartByDay"
	BlocksService_TxByHash_FullMethodName           = "/blocks.BlocksService/TxByHash"
	BlocksService_TotalTransactions_FullMethodName  = "/blocks.BlocksService/TotalTransactions"
	BlocksService_Transactions_FullMethodName       = "/blocks.BlocksService/Transactions"
	BlocksService_TotalBlocks_FullMethodName        = "/blocks.BlocksService/TotalBlocks"
	BlocksService_GetBlocks_FullMethodName          = "/blocks.BlocksService/GetBlocks"
	BlocksService_BlockSignatures_FullMethodName    = "/blocks.BlocksService/BlockSignatures"
	BlocksService_TxsByBlock_FullMethodName         = "/blocks.BlocksService/TxsByBlock"
	BlocksService_TransactionRawLog_FullMethodName  = "/blocks.BlocksService/TransactionRawLog"
	BlocksService_TransactionSigners_FullMethodName = "/blocks.BlocksService/TransactionSigners"
)

// BlocksServiceClient is the client API for BlocksService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlocksServiceClient interface {
	BlockInfo(ctx context.Context, in *GetBlockInfoRequest, opts ...grpc.CallOption) (*GetBlockInfoResponse, error)
	BlockValidators(ctx context.Context, in *GetBlockValidatorsRequest, opts ...grpc.CallOption) (*GetBlockValidatorsResponse, error)
	TxChartByDay(ctx context.Context, in *TxChartByDayRequest, opts ...grpc.CallOption) (*TxChartByDayResponse, error)
	TxByHash(ctx context.Context, in *TxByHashRequest, opts ...grpc.CallOption) (*TxByHashResponse, error)
	TotalTransactions(ctx context.Context, in *TotalTransactionsRequest, opts ...grpc.CallOption) (*TotalTransactionsResponse, error)
	Transactions(ctx context.Context, in *TransactionsRequest, opts ...grpc.CallOption) (*TransactionsResponse, error)
	TotalBlocks(ctx context.Context, in *TotalBlocksRequest, opts ...grpc.CallOption) (*TotalBlocksResponse, error)
	GetBlocks(ctx context.Context, in *GetBlocksRequest, opts ...grpc.CallOption) (*GetBlocksResponse, error)
	BlockSignatures(ctx context.Context, in *BlockSignaturesRequest, opts ...grpc.CallOption) (*BlockSignaturesResponse, error)
	TxsByBlock(ctx context.Context, in *TxsByBlockRequest, opts ...grpc.CallOption) (*TxsByBlockResponse, error)
	TransactionRawLog(ctx context.Context, in *TransactionRawLogRequest, opts ...grpc.CallOption) (*TransactionRawLogResponse, error)
	TransactionSigners(ctx context.Context, in *TransactionSignersRequest, opts ...grpc.CallOption) (*TransactionSignersResponse, error)
}

type blocksServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlocksServiceClient(cc grpc.ClientConnInterface) BlocksServiceClient {
	return &blocksServiceClient{cc}
}

func (c *blocksServiceClient) BlockInfo(ctx context.Context, in *GetBlockInfoRequest, opts ...grpc.CallOption) (*GetBlockInfoResponse, error) {
	out := new(GetBlockInfoResponse)
	err := c.cc.Invoke(ctx, BlocksService_BlockInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) BlockValidators(ctx context.Context, in *GetBlockValidatorsRequest, opts ...grpc.CallOption) (*GetBlockValidatorsResponse, error) {
	out := new(GetBlockValidatorsResponse)
	err := c.cc.Invoke(ctx, BlocksService_BlockValidators_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TxChartByDay(ctx context.Context, in *TxChartByDayRequest, opts ...grpc.CallOption) (*TxChartByDayResponse, error) {
	out := new(TxChartByDayResponse)
	err := c.cc.Invoke(ctx, BlocksService_TxChartByDay_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TxByHash(ctx context.Context, in *TxByHashRequest, opts ...grpc.CallOption) (*TxByHashResponse, error) {
	out := new(TxByHashResponse)
	err := c.cc.Invoke(ctx, BlocksService_TxByHash_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TotalTransactions(ctx context.Context, in *TotalTransactionsRequest, opts ...grpc.CallOption) (*TotalTransactionsResponse, error) {
	out := new(TotalTransactionsResponse)
	err := c.cc.Invoke(ctx, BlocksService_TotalTransactions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) Transactions(ctx context.Context, in *TransactionsRequest, opts ...grpc.CallOption) (*TransactionsResponse, error) {
	out := new(TransactionsResponse)
	err := c.cc.Invoke(ctx, BlocksService_Transactions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TotalBlocks(ctx context.Context, in *TotalBlocksRequest, opts ...grpc.CallOption) (*TotalBlocksResponse, error) {
	out := new(TotalBlocksResponse)
	err := c.cc.Invoke(ctx, BlocksService_TotalBlocks_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) GetBlocks(ctx context.Context, in *GetBlocksRequest, opts ...grpc.CallOption) (*GetBlocksResponse, error) {
	out := new(GetBlocksResponse)
	err := c.cc.Invoke(ctx, BlocksService_GetBlocks_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) BlockSignatures(ctx context.Context, in *BlockSignaturesRequest, opts ...grpc.CallOption) (*BlockSignaturesResponse, error) {
	out := new(BlockSignaturesResponse)
	err := c.cc.Invoke(ctx, BlocksService_BlockSignatures_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TxsByBlock(ctx context.Context, in *TxsByBlockRequest, opts ...grpc.CallOption) (*TxsByBlockResponse, error) {
	out := new(TxsByBlockResponse)
	err := c.cc.Invoke(ctx, BlocksService_TxsByBlock_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TransactionRawLog(ctx context.Context, in *TransactionRawLogRequest, opts ...grpc.CallOption) (*TransactionRawLogResponse, error) {
	out := new(TransactionRawLogResponse)
	err := c.cc.Invoke(ctx, BlocksService_TransactionRawLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blocksServiceClient) TransactionSigners(ctx context.Context, in *TransactionSignersRequest, opts ...grpc.CallOption) (*TransactionSignersResponse, error) {
	out := new(TransactionSignersResponse)
	err := c.cc.Invoke(ctx, BlocksService_TransactionSigners_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlocksServiceServer is the server API for BlocksService service.
// All implementations must embed UnimplementedBlocksServiceServer
// for forward compatibility
type BlocksServiceServer interface {
	BlockInfo(context.Context, *GetBlockInfoRequest) (*GetBlockInfoResponse, error)
	BlockValidators(context.Context, *GetBlockValidatorsRequest) (*GetBlockValidatorsResponse, error)
	TxChartByDay(context.Context, *TxChartByDayRequest) (*TxChartByDayResponse, error)
	TxByHash(context.Context, *TxByHashRequest) (*TxByHashResponse, error)
	TotalTransactions(context.Context, *TotalTransactionsRequest) (*TotalTransactionsResponse, error)
	Transactions(context.Context, *TransactionsRequest) (*TransactionsResponse, error)
	TotalBlocks(context.Context, *TotalBlocksRequest) (*TotalBlocksResponse, error)
	GetBlocks(context.Context, *GetBlocksRequest) (*GetBlocksResponse, error)
	BlockSignatures(context.Context, *BlockSignaturesRequest) (*BlockSignaturesResponse, error)
	TxsByBlock(context.Context, *TxsByBlockRequest) (*TxsByBlockResponse, error)
	TransactionRawLog(context.Context, *TransactionRawLogRequest) (*TransactionRawLogResponse, error)
	TransactionSigners(context.Context, *TransactionSignersRequest) (*TransactionSignersResponse, error)
	mustEmbedUnimplementedBlocksServiceServer()
}

// UnimplementedBlocksServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBlocksServiceServer struct {
}

func (UnimplementedBlocksServiceServer) BlockInfo(context.Context, *GetBlockInfoRequest) (*GetBlockInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BlockInfo not implemented")
}
func (UnimplementedBlocksServiceServer) BlockValidators(context.Context, *GetBlockValidatorsRequest) (*GetBlockValidatorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BlockValidators not implemented")
}
func (UnimplementedBlocksServiceServer) TxChartByDay(context.Context, *TxChartByDayRequest) (*TxChartByDayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxChartByDay not implemented")
}
func (UnimplementedBlocksServiceServer) TxByHash(context.Context, *TxByHashRequest) (*TxByHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxByHash not implemented")
}
func (UnimplementedBlocksServiceServer) TotalTransactions(context.Context, *TotalTransactionsRequest) (*TotalTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TotalTransactions not implemented")
}
func (UnimplementedBlocksServiceServer) Transactions(context.Context, *TransactionsRequest) (*TransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transactions not implemented")
}
func (UnimplementedBlocksServiceServer) TotalBlocks(context.Context, *TotalBlocksRequest) (*TotalBlocksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TotalBlocks not implemented")
}
func (UnimplementedBlocksServiceServer) GetBlocks(context.Context, *GetBlocksRequest) (*GetBlocksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlocks not implemented")
}
func (UnimplementedBlocksServiceServer) BlockSignatures(context.Context, *BlockSignaturesRequest) (*BlockSignaturesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BlockSignatures not implemented")
}
func (UnimplementedBlocksServiceServer) TxsByBlock(context.Context, *TxsByBlockRequest) (*TxsByBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxsByBlock not implemented")
}
func (UnimplementedBlocksServiceServer) TransactionRawLog(context.Context, *TransactionRawLogRequest) (*TransactionRawLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransactionRawLog not implemented")
}
func (UnimplementedBlocksServiceServer) TransactionSigners(context.Context, *TransactionSignersRequest) (*TransactionSignersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransactionSigners not implemented")
}
func (UnimplementedBlocksServiceServer) mustEmbedUnimplementedBlocksServiceServer() {}

// UnsafeBlocksServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlocksServiceServer will
// result in compilation errors.
type UnsafeBlocksServiceServer interface {
	mustEmbedUnimplementedBlocksServiceServer()
}

func RegisterBlocksServiceServer(s grpc.ServiceRegistrar, srv BlocksServiceServer) {
	s.RegisterService(&BlocksService_ServiceDesc, srv)
}

func _BlocksService_BlockInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).BlockInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_BlockInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).BlockInfo(ctx, req.(*GetBlockInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_BlockValidators_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockValidatorsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).BlockValidators(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_BlockValidators_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).BlockValidators(ctx, req.(*GetBlockValidatorsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TxChartByDay_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxChartByDayRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TxChartByDay(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TxChartByDay_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TxChartByDay(ctx, req.(*TxChartByDayRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TxByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxByHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TxByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TxByHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TxByHash(ctx, req.(*TxByHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TotalTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TotalTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TotalTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TotalTransactions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TotalTransactions(ctx, req.(*TotalTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_Transactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).Transactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_Transactions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).Transactions(ctx, req.(*TransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TotalBlocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TotalBlocksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TotalBlocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TotalBlocks_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TotalBlocks(ctx, req.(*TotalBlocksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_GetBlocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlocksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).GetBlocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_GetBlocks_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).GetBlocks(ctx, req.(*GetBlocksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_BlockSignatures_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockSignaturesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).BlockSignatures(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_BlockSignatures_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).BlockSignatures(ctx, req.(*BlockSignaturesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TxsByBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxsByBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TxsByBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TxsByBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TxsByBlock(ctx, req.(*TxsByBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TransactionRawLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionRawLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TransactionRawLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TransactionRawLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TransactionRawLog(ctx, req.(*TransactionRawLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlocksService_TransactionSigners_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionSignersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).TransactionSigners(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_TransactionSigners_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).TransactionSigners(ctx, req.(*TransactionSignersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BlocksService_ServiceDesc is the grpc.ServiceDesc for BlocksService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BlocksService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "blocks.BlocksService",
	HandlerType: (*BlocksServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BlockInfo",
			Handler:    _BlocksService_BlockInfo_Handler,
		},
		{
			MethodName: "BlockValidators",
			Handler:    _BlocksService_BlockValidators_Handler,
		},
		{
			MethodName: "TxChartByDay",
			Handler:    _BlocksService_TxChartByDay_Handler,
		},
		{
			MethodName: "TxByHash",
			Handler:    _BlocksService_TxByHash_Handler,
		},
		{
			MethodName: "TotalTransactions",
			Handler:    _BlocksService_TotalTransactions_Handler,
		},
		{
			MethodName: "Transactions",
			Handler:    _BlocksService_Transactions_Handler,
		},
		{
			MethodName: "TotalBlocks",
			Handler:    _BlocksService_TotalBlocks_Handler,
		},
		{
			MethodName: "GetBlocks",
			Handler:    _BlocksService_GetBlocks_Handler,
		},
		{
			MethodName: "BlockSignatures",
			Handler:    _BlocksService_BlockSignatures_Handler,
		},
		{
			MethodName: "TxsByBlock",
			Handler:    _BlocksService_TxsByBlock_Handler,
		},
		{
			MethodName: "TransactionRawLog",
			Handler:    _BlocksService_TransactionRawLog_Handler,
		},
		{
			MethodName: "TransactionSigners",
			Handler:    _BlocksService_TransactionSigners_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "blocks.proto",
}
