// Code generated by protoc-gen-go. DO NOT EDIT.
// source: backend/user/user.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	backend/user/user.proto

It has these top-level messages:
	GetUserReq
	GetUserByInternalAddressReq
	UserInfo
	IsExistReq
	IsExistResp
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type GetUserReq struct {
	UserId uint64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *GetUserReq) Reset()                    { *m = GetUserReq{} }
func (m *GetUserReq) String() string            { return proto1.CompactTextString(m) }
func (*GetUserReq) ProtoMessage()               {}
func (*GetUserReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *GetUserReq) GetUserId() uint64 {
	if m != nil {
		return m.UserId
	}
	return 0
}

type GetUserByInternalAddressReq struct {
	InternalAddress string `protobuf:"bytes,1,opt,name=internal_address,json=internalAddress" json:"internal_address,omitempty"`
}

func (m *GetUserByInternalAddressReq) Reset()                    { *m = GetUserByInternalAddressReq{} }
func (m *GetUserByInternalAddressReq) String() string            { return proto1.CompactTextString(m) }
func (*GetUserByInternalAddressReq) ProtoMessage()               {}
func (*GetUserByInternalAddressReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *GetUserByInternalAddressReq) GetInternalAddress() string {
	if m != nil {
		return m.InternalAddress
	}
	return ""
}

type UserInfo struct {
	Id              uint64  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name            string  `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	AreaCode        string  `protobuf:"bytes,3,opt,name=area_code,json=areaCode" json:"area_code,omitempty"`
	Mobile          string  `protobuf:"bytes,4,opt,name=mobile" json:"mobile,omitempty"`
	Email           string  `protobuf:"bytes,5,opt,name=email" json:"email,omitempty"`
	AvatarUrl       string  `protobuf:"bytes,6,opt,name=avatar_url,json=avatarUrl" json:"avatar_url,omitempty"`
	Status          int32   `protobuf:"varint,7,opt,name=status" json:"status,omitempty"`
	InternalAddress string  `protobuf:"bytes,8,opt,name=internal_address,json=internalAddress" json:"internal_address,omitempty"`
	ParentId        uint64  `protobuf:"varint,9,opt,name=parent_id,json=parentId" json:"parent_id,omitempty"`
	UserType        int32   `protobuf:"varint,10,opt,name=user_type,json=userType" json:"user_type,omitempty"`
	DiscountRate    float64 `protobuf:"fixed64,11,opt,name=discount_rate,json=discountRate" json:"discount_rate,omitempty"`
	IsRealName      bool    `protobuf:"varint,12,opt,name=is_real_name,json=isRealName" json:"is_real_name,omitempty"`
	ReferralCode    string  `protobuf:"bytes,13,opt,name=referral_code,json=referralCode" json:"referral_code,omitempty"`
	TradingMethods  string  `protobuf:"bytes,14,opt,name=trading_methods,json=tradingMethods" json:"trading_methods,omitempty"`
}

func (m *UserInfo) Reset()                    { *m = UserInfo{} }
func (m *UserInfo) String() string            { return proto1.CompactTextString(m) }
func (*UserInfo) ProtoMessage()               {}
func (*UserInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *UserInfo) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *UserInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserInfo) GetAreaCode() string {
	if m != nil {
		return m.AreaCode
	}
	return ""
}

func (m *UserInfo) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

func (m *UserInfo) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *UserInfo) GetAvatarUrl() string {
	if m != nil {
		return m.AvatarUrl
	}
	return ""
}

func (m *UserInfo) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *UserInfo) GetInternalAddress() string {
	if m != nil {
		return m.InternalAddress
	}
	return ""
}

func (m *UserInfo) GetParentId() uint64 {
	if m != nil {
		return m.ParentId
	}
	return 0
}

func (m *UserInfo) GetUserType() int32 {
	if m != nil {
		return m.UserType
	}
	return 0
}

func (m *UserInfo) GetDiscountRate() float64 {
	if m != nil {
		return m.DiscountRate
	}
	return 0
}

func (m *UserInfo) GetIsRealName() bool {
	if m != nil {
		return m.IsRealName
	}
	return false
}

func (m *UserInfo) GetReferralCode() string {
	if m != nil {
		return m.ReferralCode
	}
	return ""
}

func (m *UserInfo) GetTradingMethods() string {
	if m != nil {
		return m.TradingMethods
	}
	return ""
}

type IsExistReq struct {
	Input string `protobuf:"bytes,1,opt,name=input" json:"input,omitempty"`
}

func (m *IsExistReq) Reset()                    { *m = IsExistReq{} }
func (m *IsExistReq) String() string            { return proto1.CompactTextString(m) }
func (*IsExistReq) ProtoMessage()               {}
func (*IsExistReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *IsExistReq) GetInput() string {
	if m != nil {
		return m.Input
	}
	return ""
}

type IsExistResp struct {
	Exist bool `protobuf:"varint,1,opt,name=exist" json:"exist,omitempty"`
}

func (m *IsExistResp) Reset()                    { *m = IsExistResp{} }
func (m *IsExistResp) String() string            { return proto1.CompactTextString(m) }
func (*IsExistResp) ProtoMessage()               {}
func (*IsExistResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *IsExistResp) GetExist() bool {
	if m != nil {
		return m.Exist
	}
	return false
}

func init() {
	proto1.RegisterType((*GetUserReq)(nil), "proto.GetUserReq")
	proto1.RegisterType((*GetUserByInternalAddressReq)(nil), "proto.GetUserByInternalAddressReq")
	proto1.RegisterType((*UserInfo)(nil), "proto.UserInfo")
	proto1.RegisterType((*IsExistReq)(nil), "proto.IsExistReq")
	proto1.RegisterType((*IsExistResp)(nil), "proto.IsExistResp")
}

func init() { proto1.RegisterFile("backend/user/user.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 472 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x5f, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xeb, 0x34, 0x4e, 0x9c, 0xc9, 0x3f, 0x58, 0x55, 0x74, 0xd5, 0x0a, 0xc9, 0x72, 0x84,
	0x08, 0x2f, 0xad, 0x54, 0x4e, 0x00, 0x08, 0x81, 0x1f, 0x40, 0xc2, 0xd0, 0x67, 0x6b, 0x92, 0x9d,
	0xc2, 0x0a, 0xc7, 0x36, 0xbb, 0xeb, 0x8a, 0x5c, 0x89, 0x7b, 0x70, 0x2f, 0xb4, 0x63, 0x27, 0x55,
	0x50, 0xfa, 0x92, 0xf8, 0xfb, 0xcd, 0xec, 0xec, 0xcc, 0xb7, 0x03, 0xe7, 0x2b, 0x5c, 0xff, 0xa4,
	0x52, 0x5d, 0x37, 0x96, 0x0c, 0xff, 0x5c, 0xd5, 0xa6, 0x72, 0x95, 0x08, 0xf9, 0x2f, 0x79, 0x01,
	0xf0, 0x81, 0xdc, 0xad, 0x25, 0x93, 0xd1, 0x2f, 0x71, 0x0e, 0x43, 0x9f, 0x92, 0x6b, 0x25, 0x83,
	0x38, 0x58, 0xf6, 0xb3, 0x81, 0x97, 0xa9, 0x4a, 0x3e, 0xc2, 0x65, 0x97, 0xf6, 0x76, 0x9b, 0x96,
	0x8e, 0x4c, 0x89, 0xc5, 0x1b, 0xa5, 0x0c, 0x59, 0xeb, 0xcf, 0xbd, 0x82, 0x27, 0xba, 0xa3, 0x39,
	0xb6, 0x98, 0x0b, 0x8c, 0xb2, 0xb9, 0x3e, 0xcc, 0x4e, 0xfe, 0x9c, 0x42, 0xe4, 0xeb, 0xa4, 0xe5,
	0x5d, 0x25, 0x66, 0xd0, 0xdb, 0x5f, 0xd5, 0xd3, 0x4a, 0x08, 0xe8, 0x97, 0xb8, 0x21, 0xd9, 0xe3,
	0xb3, 0xfc, 0x2d, 0x2e, 0x61, 0x84, 0x86, 0x30, 0x5f, 0x57, 0x8a, 0xe4, 0x29, 0x07, 0x22, 0x0f,
	0xde, 0x55, 0x8a, 0xc4, 0x33, 0x18, 0x6c, 0xaa, 0x95, 0x2e, 0x48, 0xf6, 0x39, 0xd2, 0x29, 0x71,
	0x06, 0x21, 0x6d, 0x50, 0x17, 0x32, 0x64, 0xdc, 0x0a, 0xf1, 0x1c, 0x00, 0xef, 0xd1, 0xa1, 0xc9,
	0x1b, 0x53, 0xc8, 0x01, 0x87, 0x46, 0x2d, 0xb9, 0x35, 0x85, 0x2f, 0x66, 0x1d, 0xba, 0xc6, 0xca,
	0x61, 0x1c, 0x2c, 0xc3, 0xac, 0x53, 0x47, 0xa7, 0x8b, 0x8e, 0x4e, 0xe7, 0x9b, 0xad, 0xd1, 0x50,
	0xe9, 0xbc, 0x85, 0x23, 0x9e, 0x2b, 0x6a, 0x41, 0xaa, 0x7c, 0x90, 0xdd, 0x75, 0xdb, 0x9a, 0x24,
	0xf0, 0x15, 0x91, 0x07, 0xdf, 0xb6, 0x35, 0x89, 0x05, 0x4c, 0x95, 0xb6, 0xeb, 0xaa, 0x29, 0x5d,
	0x6e, 0xd0, 0x91, 0x1c, 0xc7, 0xc1, 0x32, 0xc8, 0x26, 0x3b, 0x98, 0xa1, 0x23, 0x11, 0xc3, 0x44,
	0xdb, 0xdc, 0x10, 0x16, 0x39, 0xfb, 0x34, 0x89, 0x83, 0x65, 0x94, 0x81, 0xb6, 0x19, 0x61, 0xf1,
	0xd9, 0xbb, 0xb5, 0x80, 0xa9, 0xa1, 0x3b, 0x32, 0x06, 0x8b, 0xd6, 0xb1, 0x29, 0x37, 0x3a, 0xd9,
	0x41, 0x76, 0xed, 0x25, 0xcc, 0x9d, 0x41, 0xa5, 0xcb, 0xef, 0xf9, 0x86, 0xdc, 0x8f, 0x4a, 0x59,
	0x39, 0xe3, 0xb4, 0x59, 0x87, 0x3f, 0xb5, 0x34, 0x49, 0x00, 0x52, 0xfb, 0xfe, 0xb7, 0xb6, 0xce,
	0xbf, 0xf2, 0x19, 0x84, 0xba, 0xac, 0x1b, 0xd7, 0x3d, 0x6d, 0x2b, 0x92, 0x05, 0x8c, 0xf7, 0x39,
	0xb6, 0x66, 0xe7, 0xbd, 0xe0, 0xa4, 0x28, 0x6b, 0xc5, 0xcd, 0xdf, 0x00, 0xc6, 0xfe, 0xd5, 0xbf,
	0x92, 0xb9, 0xd7, 0x6b, 0x12, 0xd7, 0x30, 0xec, 0xf6, 0x49, 0x3c, 0x6d, 0x17, 0xf2, 0xea, 0x61,
	0x0d, 0x2f, 0xe6, 0x1d, 0xda, 0xed, 0x49, 0x72, 0x22, 0xbe, 0x80, 0x7c, 0x6c, 0x01, 0x45, 0x72,
	0x58, 0xe1, 0xd8, 0x86, 0x1e, 0x2b, 0x79, 0x03, 0xc3, 0xae, 0xf1, 0x7d, 0x0f, 0x0f, 0xc3, 0x5e,
	0x88, 0xff, 0x91, 0xad, 0x93, 0x93, 0xd5, 0x80, 0xe1, 0xeb, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x06, 0xd8, 0xb6, 0xf2, 0x57, 0x03, 0x00, 0x00,
}