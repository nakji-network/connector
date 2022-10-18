// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: cCOMP/cCOMP.proto

package cCOMP

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Mint represents a Mint event raised by the Compound contract.
type Mint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts         *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Block      uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	Idx        uint64                 `protobuf:"varint,3,opt,name=idx,proto3" json:"idx,omitempty"`
	Tx         []byte                 `protobuf:"bytes,4,opt,name=tx,proto3" json:"tx,omitempty"`         // tx hash
	Minter     []byte                 `protobuf:"bytes,5,opt,name=minter,proto3" json:"minter,omitempty"` // The address that minted the assets
	MintAmount []byte                 `protobuf:"bytes,6,opt,name=mintAmount,proto3" json:"mintAmount,omitempty"`
	MintTokens []byte                 `protobuf:"bytes,7,opt,name=mintTokens,proto3" json:"mintTokens,omitempty"`
}

func (x *Mint) Reset() {
	*x = Mint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cCOMP_cCOMP_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Mint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Mint) ProtoMessage() {}

func (x *Mint) ProtoReflect() protoreflect.Message {
	mi := &file_cCOMP_cCOMP_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Mint.ProtoReflect.Descriptor instead.
func (*Mint) Descriptor() ([]byte, []int) {
	return file_cCOMP_cCOMP_proto_rawDescGZIP(), []int{0}
}

func (x *Mint) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *Mint) GetBlock() uint64 {
	if x != nil {
		return x.Block
	}
	return 0
}

func (x *Mint) GetIdx() uint64 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *Mint) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *Mint) GetMinter() []byte {
	if x != nil {
		return x.Minter
	}
	return nil
}

func (x *Mint) GetMintAmount() []byte {
	if x != nil {
		return x.MintAmount
	}
	return nil
}

func (x *Mint) GetMintTokens() []byte {
	if x != nil {
		return x.MintTokens
	}
	return nil
}

// Redeem represents a Borrow event raised by the Compound contract.
type Redeem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts           *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Block        uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	Idx          uint64                 `protobuf:"varint,3,opt,name=idx,proto3" json:"idx,omitempty"`
	Tx           []byte                 `protobuf:"bytes,4,opt,name=tx,proto3" json:"tx,omitempty"`             // tx hash
	Redeemer     []byte                 `protobuf:"bytes,5,opt,name=redeemer,proto3" json:"redeemer,omitempty"` // The address that redeemed the assets
	RedeemAmount []byte                 `protobuf:"bytes,6,opt,name=redeemAmount,proto3" json:"redeemAmount,omitempty"`
	RedeemTokens []byte                 `protobuf:"bytes,7,opt,name=redeemTokens,proto3" json:"redeemTokens,omitempty"`
}

func (x *Redeem) Reset() {
	*x = Redeem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cCOMP_cCOMP_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Redeem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Redeem) ProtoMessage() {}

func (x *Redeem) ProtoReflect() protoreflect.Message {
	mi := &file_cCOMP_cCOMP_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Redeem.ProtoReflect.Descriptor instead.
func (*Redeem) Descriptor() ([]byte, []int) {
	return file_cCOMP_cCOMP_proto_rawDescGZIP(), []int{1}
}

func (x *Redeem) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *Redeem) GetBlock() uint64 {
	if x != nil {
		return x.Block
	}
	return 0
}

func (x *Redeem) GetIdx() uint64 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *Redeem) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *Redeem) GetRedeemer() []byte {
	if x != nil {
		return x.Redeemer
	}
	return nil
}

func (x *Redeem) GetRedeemAmount() []byte {
	if x != nil {
		return x.RedeemAmount
	}
	return nil
}

func (x *Redeem) GetRedeemTokens() []byte {
	if x != nil {
		return x.RedeemTokens
	}
	return nil
}

// Borrow represents a Borrow event raised by the Compound contract.
type Borrow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts             *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Block          uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	Idx            uint64                 `protobuf:"varint,3,opt,name=idx,proto3" json:"idx,omitempty"`
	Tx             []byte                 `protobuf:"bytes,4,opt,name=tx,proto3" json:"tx,omitempty"`             // tx hash
	Borrower       []byte                 `protobuf:"bytes,5,opt,name=borrower,proto3" json:"borrower,omitempty"` // The address that borrowed the assets
	BorrowAmount   []byte                 `protobuf:"bytes,6,opt,name=borrowAmount,proto3" json:"borrowAmount,omitempty"`
	AccountBorrows []byte                 `protobuf:"bytes,7,opt,name=accountBorrows,proto3" json:"accountBorrows,omitempty"`
	TotalBorrows   []byte                 `protobuf:"bytes,8,opt,name=totalBorrows,proto3" json:"totalBorrows,omitempty"`
}

func (x *Borrow) Reset() {
	*x = Borrow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cCOMP_cCOMP_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Borrow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Borrow) ProtoMessage() {}

func (x *Borrow) ProtoReflect() protoreflect.Message {
	mi := &file_cCOMP_cCOMP_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Borrow.ProtoReflect.Descriptor instead.
func (*Borrow) Descriptor() ([]byte, []int) {
	return file_cCOMP_cCOMP_proto_rawDescGZIP(), []int{2}
}

func (x *Borrow) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *Borrow) GetBlock() uint64 {
	if x != nil {
		return x.Block
	}
	return 0
}

func (x *Borrow) GetIdx() uint64 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *Borrow) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *Borrow) GetBorrower() []byte {
	if x != nil {
		return x.Borrower
	}
	return nil
}

func (x *Borrow) GetBorrowAmount() []byte {
	if x != nil {
		return x.BorrowAmount
	}
	return nil
}

func (x *Borrow) GetAccountBorrows() []byte {
	if x != nil {
		return x.AccountBorrows
	}
	return nil
}

func (x *Borrow) GetTotalBorrows() []byte {
	if x != nil {
		return x.TotalBorrows
	}
	return nil
}

// RepayBorrow represents a RepayBorrow event raised by the Compound contract.
type RepayBorrow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts             *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Block          uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	Idx            uint64                 `protobuf:"varint,3,opt,name=idx,proto3" json:"idx,omitempty"`
	Tx             []byte                 `protobuf:"bytes,4,opt,name=tx,proto3" json:"tx,omitempty"` // tx hash
	Payer          []byte                 `protobuf:"bytes,5,opt,name=payer,proto3" json:"payer,omitempty"`
	Borrower       []byte                 `protobuf:"bytes,6,opt,name=borrower,proto3" json:"borrower,omitempty"`
	RepayAmount    []byte                 `protobuf:"bytes,7,opt,name=repayAmount,proto3" json:"repayAmount,omitempty"`
	AccountBorrows []byte                 `protobuf:"bytes,8,opt,name=accountBorrows,proto3" json:"accountBorrows,omitempty"`
	TotalBorrows   []byte                 `protobuf:"bytes,9,opt,name=totalBorrows,proto3" json:"totalBorrows,omitempty"`
}

func (x *RepayBorrow) Reset() {
	*x = RepayBorrow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cCOMP_cCOMP_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepayBorrow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepayBorrow) ProtoMessage() {}

func (x *RepayBorrow) ProtoReflect() protoreflect.Message {
	mi := &file_cCOMP_cCOMP_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepayBorrow.ProtoReflect.Descriptor instead.
func (*RepayBorrow) Descriptor() ([]byte, []int) {
	return file_cCOMP_cCOMP_proto_rawDescGZIP(), []int{3}
}

func (x *RepayBorrow) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *RepayBorrow) GetBlock() uint64 {
	if x != nil {
		return x.Block
	}
	return 0
}

func (x *RepayBorrow) GetIdx() uint64 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *RepayBorrow) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *RepayBorrow) GetPayer() []byte {
	if x != nil {
		return x.Payer
	}
	return nil
}

func (x *RepayBorrow) GetBorrower() []byte {
	if x != nil {
		return x.Borrower
	}
	return nil
}

func (x *RepayBorrow) GetRepayAmount() []byte {
	if x != nil {
		return x.RepayAmount
	}
	return nil
}

func (x *RepayBorrow) GetAccountBorrows() []byte {
	if x != nil {
		return x.AccountBorrows
	}
	return nil
}

func (x *RepayBorrow) GetTotalBorrows() []byte {
	if x != nil {
		return x.TotalBorrows
	}
	return nil
}

// LiquidateBorrow represents a LiquidateBorrow event raised by the Compound contract.
type LiquidateBorrow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ts               *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=ts,proto3" json:"ts,omitempty"`
	Block            uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	Idx              uint64                 `protobuf:"varint,3,opt,name=idx,proto3" json:"idx,omitempty"`
	Tx               []byte                 `protobuf:"bytes,4,opt,name=tx,proto3" json:"tx,omitempty"` // tx hash
	Liquidator       []byte                 `protobuf:"bytes,5,opt,name=liquidator,proto3" json:"liquidator,omitempty"`
	Borrower         []byte                 `protobuf:"bytes,6,opt,name=borrower,proto3" json:"borrower,omitempty"`
	RepayAmount      []byte                 `protobuf:"bytes,7,opt,name=repayAmount,proto3" json:"repayAmount,omitempty"`
	CTokenCollateral []byte                 `protobuf:"bytes,8,opt,name=cTokenCollateral,proto3" json:"cTokenCollateral,omitempty"`
	SeizeTokens      []byte                 `protobuf:"bytes,9,opt,name=seizeTokens,proto3" json:"seizeTokens,omitempty"`
}

func (x *LiquidateBorrow) Reset() {
	*x = LiquidateBorrow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cCOMP_cCOMP_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LiquidateBorrow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LiquidateBorrow) ProtoMessage() {}

func (x *LiquidateBorrow) ProtoReflect() protoreflect.Message {
	mi := &file_cCOMP_cCOMP_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LiquidateBorrow.ProtoReflect.Descriptor instead.
func (*LiquidateBorrow) Descriptor() ([]byte, []int) {
	return file_cCOMP_cCOMP_proto_rawDescGZIP(), []int{4}
}

func (x *LiquidateBorrow) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *LiquidateBorrow) GetBlock() uint64 {
	if x != nil {
		return x.Block
	}
	return 0
}

func (x *LiquidateBorrow) GetIdx() uint64 {
	if x != nil {
		return x.Idx
	}
	return 0
}

func (x *LiquidateBorrow) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *LiquidateBorrow) GetLiquidator() []byte {
	if x != nil {
		return x.Liquidator
	}
	return nil
}

func (x *LiquidateBorrow) GetBorrower() []byte {
	if x != nil {
		return x.Borrower
	}
	return nil
}

func (x *LiquidateBorrow) GetRepayAmount() []byte {
	if x != nil {
		return x.RepayAmount
	}
	return nil
}

func (x *LiquidateBorrow) GetCTokenCollateral() []byte {
	if x != nil {
		return x.CTokenCollateral
	}
	return nil
}

func (x *LiquidateBorrow) GetSeizeTokens() []byte {
	if x != nil {
		return x.SeizeTokens
	}
	return nil
}

var File_cCOMP_cCOMP_proto protoreflect.FileDescriptor

var file_cCOMP_cCOMP_proto_rawDesc = []byte{
	0x0a, 0x11, 0x63, 0x43, 0x4f, 0x4d, 0x50, 0x2f, 0x63, 0x43, 0x4f, 0x4d, 0x50, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x63, 0x43, 0x4f, 0x4d, 0x50, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc2, 0x01, 0x0a, 0x04,
	0x4d, 0x69, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x78, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x03, 0x69, 0x64, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x12, 0x1e, 0x0a, 0x0a, 0x6d, 0x69, 0x6e, 0x74, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x6d, 0x69, 0x6e, 0x74, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x1e, 0x0a, 0x0a, 0x6d, 0x69, 0x6e, 0x74, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x6d, 0x69, 0x6e, 0x74, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73,
	0x22, 0xd0, 0x01, 0x0a, 0x06, 0x52, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x12, 0x2a, 0x0a, 0x02, 0x74,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x10, 0x0a,
	0x03, 0x69, 0x64, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x69, 0x64, 0x78, 0x12,
	0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78, 0x12,
	0x1a, 0x0a, 0x08, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x08, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x65, 0x72, 0x12, 0x22, 0x0a, 0x0c, 0x72,
	0x65, 0x64, 0x65, 0x65, 0x6d, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0c, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x22, 0x0a, 0x0c, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x72, 0x65, 0x64, 0x65, 0x65, 0x6d, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x73, 0x22, 0xf8, 0x01, 0x0a, 0x06, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x12, 0x2a,
	0x0a, 0x02, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x69,
	0x64, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02,
	0x74, 0x78, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x65, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x62, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x22,
	0x0a, 0x0c, 0x62, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x62, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x41, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x42, 0x6f, 0x72,
	0x72, 0x6f, 0x77, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0c, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x22, 0x91,
	0x02, 0x0a, 0x0b, 0x52, 0x65, 0x70, 0x61, 0x79, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x12, 0x2a,
	0x0a, 0x02, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x69,
	0x64, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02,
	0x74, 0x78, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61, 0x79, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x70, 0x61, 0x79, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x6f, 0x72, 0x72,
	0x6f, 0x77, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x62, 0x6f, 0x72, 0x72,
	0x6f, 0x77, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x61, 0x79, 0x41, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x61, 0x79,
	0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x12, 0x22,
	0x0a, 0x0c, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x42, 0x6f, 0x72, 0x72, 0x6f,
	0x77, 0x73, 0x22, 0xa1, 0x02, 0x0a, 0x0f, 0x4c, 0x69, 0x71, 0x75, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x42, 0x6f, 0x72, 0x72, 0x6f, 0x77, 0x12, 0x2a, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02,
	0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x78, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x69, 0x64, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78, 0x12, 0x1e, 0x0a, 0x0a, 0x6c, 0x69,
	0x71, 0x75, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a,
	0x6c, 0x69, 0x71, 0x75, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x62, 0x6f,
	0x72, 0x72, 0x6f, 0x77, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x62, 0x6f,
	0x72, 0x72, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x61, 0x79, 0x41,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x72, 0x65, 0x70,
	0x61, 0x79, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x10, 0x63, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x43, 0x6f, 0x6c, 0x6c, 0x61, 0x74, 0x65, 0x72, 0x61, 0x6c, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x10, 0x63, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x43, 0x6f, 0x6c, 0x6c, 0x61, 0x74,
	0x65, 0x72, 0x61, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x65, 0x69, 0x7a, 0x65, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x65, 0x69, 0x7a, 0x65,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x61, 0x6b, 0x6a, 0x69, 0x2d, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x65, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x75, 0x6e, 0x64, 0x2f, 0x63,
	0x43, 0x4f, 0x4d, 0x50, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cCOMP_cCOMP_proto_rawDescOnce sync.Once
	file_cCOMP_cCOMP_proto_rawDescData = file_cCOMP_cCOMP_proto_rawDesc
)

func file_cCOMP_cCOMP_proto_rawDescGZIP() []byte {
	file_cCOMP_cCOMP_proto_rawDescOnce.Do(func() {
		file_cCOMP_cCOMP_proto_rawDescData = protoimpl.X.CompressGZIP(file_cCOMP_cCOMP_proto_rawDescData)
	})
	return file_cCOMP_cCOMP_proto_rawDescData
}

var file_cCOMP_cCOMP_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_cCOMP_cCOMP_proto_goTypes = []interface{}{
	(*Mint)(nil),                  // 0: cCOMP.Mint
	(*Redeem)(nil),                // 1: cCOMP.Redeem
	(*Borrow)(nil),                // 2: cCOMP.Borrow
	(*RepayBorrow)(nil),           // 3: cCOMP.RepayBorrow
	(*LiquidateBorrow)(nil),       // 4: cCOMP.LiquidateBorrow
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_cCOMP_cCOMP_proto_depIdxs = []int32{
	5, // 0: cCOMP.Mint.ts:type_name -> google.protobuf.Timestamp
	5, // 1: cCOMP.Redeem.ts:type_name -> google.protobuf.Timestamp
	5, // 2: cCOMP.Borrow.ts:type_name -> google.protobuf.Timestamp
	5, // 3: cCOMP.RepayBorrow.ts:type_name -> google.protobuf.Timestamp
	5, // 4: cCOMP.LiquidateBorrow.ts:type_name -> google.protobuf.Timestamp
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_cCOMP_cCOMP_proto_init() }
func file_cCOMP_cCOMP_proto_init() {
	if File_cCOMP_cCOMP_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cCOMP_cCOMP_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Mint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cCOMP_cCOMP_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Redeem); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cCOMP_cCOMP_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Borrow); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cCOMP_cCOMP_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepayBorrow); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cCOMP_cCOMP_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LiquidateBorrow); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cCOMP_cCOMP_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cCOMP_cCOMP_proto_goTypes,
		DependencyIndexes: file_cCOMP_cCOMP_proto_depIdxs,
		MessageInfos:      file_cCOMP_cCOMP_proto_msgTypes,
	}.Build()
	File_cCOMP_cCOMP_proto = out.File
	file_cCOMP_cCOMP_proto_rawDesc = nil
	file_cCOMP_cCOMP_proto_goTypes = nil
	file_cCOMP_cCOMP_proto_depIdxs = nil
}
