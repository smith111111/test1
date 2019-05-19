// Code generated by protoc-gen-go. DO NOT EDIT.
// source: im/im.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	im/im.proto

It has these top-level messages:
	UserRegisterReq
	UserRegisterResp
	UserUpdateReq
	UserUpdateResp
	SendSysMsgReq
	SendSysMsgResp
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

type UserRegisterReq struct {
	Id     uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Props  []byte `protobuf:"bytes,3,opt,name=props,proto3" json:"props,omitempty"`
	Icon   string `protobuf:"bytes,4,opt,name=icon" json:"icon,omitempty"`
	Email  string `protobuf:"bytes,5,opt,name=email" json:"email,omitempty"`
	Birth  string `protobuf:"bytes,6,opt,name=birth" json:"birth,omitempty"`
	Mobile string `protobuf:"bytes,7,opt,name=mobile" json:"mobile,omitempty"`
	Gender int32  `protobuf:"varint,8,opt,name=gender" json:"gender,omitempty"`
	Ex     []byte `protobuf:"bytes,9,opt,name=ex,proto3" json:"ex,omitempty"`
}

func (m *UserRegisterReq) Reset()                    { *m = UserRegisterReq{} }
func (m *UserRegisterReq) String() string            { return proto1.CompactTextString(m) }
func (*UserRegisterReq) ProtoMessage()               {}
func (*UserRegisterReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *UserRegisterReq) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *UserRegisterReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserRegisterReq) GetProps() []byte {
	if m != nil {
		return m.Props
	}
	return nil
}

func (m *UserRegisterReq) GetIcon() string {
	if m != nil {
		return m.Icon
	}
	return ""
}

func (m *UserRegisterReq) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *UserRegisterReq) GetBirth() string {
	if m != nil {
		return m.Birth
	}
	return ""
}

func (m *UserRegisterReq) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

func (m *UserRegisterReq) GetGender() int32 {
	if m != nil {
		return m.Gender
	}
	return 0
}

func (m *UserRegisterReq) GetEx() []byte {
	if m != nil {
		return m.Ex
	}
	return nil
}

type UserRegisterResp struct {
	ErrNo int32  `protobuf:"varint,1,opt,name=errNo" json:"errNo,omitempty"`
	Msg   string `protobuf:"bytes,2,opt,name=msg" json:"msg,omitempty"`
	Token string `protobuf:"bytes,3,opt,name=token" json:"token,omitempty"`
}

func (m *UserRegisterResp) Reset()                    { *m = UserRegisterResp{} }
func (m *UserRegisterResp) String() string            { return proto1.CompactTextString(m) }
func (*UserRegisterResp) ProtoMessage()               {}
func (*UserRegisterResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *UserRegisterResp) GetErrNo() int32 {
	if m != nil {
		return m.ErrNo
	}
	return 0
}

func (m *UserRegisterResp) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *UserRegisterResp) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type UserUpdateReq struct {
	Id     uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Icon   string `protobuf:"bytes,4,opt,name=icon" json:"icon,omitempty"`
	Email  string `protobuf:"bytes,5,opt,name=email" json:"email,omitempty"`
	Birth  string `protobuf:"bytes,6,opt,name=birth" json:"birth,omitempty"`
	Mobile string `protobuf:"bytes,7,opt,name=mobile" json:"mobile,omitempty"`
	Gender int32  `protobuf:"varint,8,opt,name=gender" json:"gender,omitempty"`
	Ex     []byte `protobuf:"bytes,9,opt,name=ex,proto3" json:"ex,omitempty"`
}

func (m *UserUpdateReq) Reset()                    { *m = UserUpdateReq{} }
func (m *UserUpdateReq) String() string            { return proto1.CompactTextString(m) }
func (*UserUpdateReq) ProtoMessage()               {}
func (*UserUpdateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *UserUpdateReq) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *UserUpdateReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserUpdateReq) GetIcon() string {
	if m != nil {
		return m.Icon
	}
	return ""
}

func (m *UserUpdateReq) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *UserUpdateReq) GetBirth() string {
	if m != nil {
		return m.Birth
	}
	return ""
}

func (m *UserUpdateReq) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

func (m *UserUpdateReq) GetGender() int32 {
	if m != nil {
		return m.Gender
	}
	return 0
}

func (m *UserUpdateReq) GetEx() []byte {
	if m != nil {
		return m.Ex
	}
	return nil
}

type UserUpdateResp struct {
	ErrNo int32  `protobuf:"varint,1,opt,name=errNo" json:"errNo,omitempty"`
	Msg   string `protobuf:"bytes,2,opt,name=msg" json:"msg,omitempty"`
}

func (m *UserUpdateResp) Reset()                    { *m = UserUpdateResp{} }
func (m *UserUpdateResp) String() string            { return proto1.CompactTextString(m) }
func (*UserUpdateResp) ProtoMessage()               {}
func (*UserUpdateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *UserUpdateResp) GetErrNo() int32 {
	if m != nil {
		return m.ErrNo
	}
	return 0
}

func (m *UserUpdateResp) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type SendSysMsgReq struct {
	From   uint64 `protobuf:"varint,1,opt,name=from" json:"from,omitempty"`
	To     uint64 `protobuf:"varint,2,opt,name=to" json:"to,omitempty"`
	Attach []byte `protobuf:"bytes,3,opt,name=attach,proto3" json:"attach,omitempty"`
}

func (m *SendSysMsgReq) Reset()                    { *m = SendSysMsgReq{} }
func (m *SendSysMsgReq) String() string            { return proto1.CompactTextString(m) }
func (*SendSysMsgReq) ProtoMessage()               {}
func (*SendSysMsgReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *SendSysMsgReq) GetFrom() uint64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *SendSysMsgReq) GetTo() uint64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *SendSysMsgReq) GetAttach() []byte {
	if m != nil {
		return m.Attach
	}
	return nil
}

type SendSysMsgResp struct {
	ErrNo int32  `protobuf:"varint,1,opt,name=errNo" json:"errNo,omitempty"`
	Msg   string `protobuf:"bytes,2,opt,name=msg" json:"msg,omitempty"`
}

func (m *SendSysMsgResp) Reset()                    { *m = SendSysMsgResp{} }
func (m *SendSysMsgResp) String() string            { return proto1.CompactTextString(m) }
func (*SendSysMsgResp) ProtoMessage()               {}
func (*SendSysMsgResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *SendSysMsgResp) GetErrNo() int32 {
	if m != nil {
		return m.ErrNo
	}
	return 0
}

func (m *SendSysMsgResp) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func init() {
	proto1.RegisterType((*UserRegisterReq)(nil), "proto.UserRegisterReq")
	proto1.RegisterType((*UserRegisterResp)(nil), "proto.UserRegisterResp")
	proto1.RegisterType((*UserUpdateReq)(nil), "proto.UserUpdateReq")
	proto1.RegisterType((*UserUpdateResp)(nil), "proto.UserUpdateResp")
	proto1.RegisterType((*SendSysMsgReq)(nil), "proto.SendSysMsgReq")
	proto1.RegisterType((*SendSysMsgResp)(nil), "proto.SendSysMsgResp")
}

func init() { proto1.RegisterFile("im/im.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 370 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x52, 0xc1, 0x6e, 0xe2, 0x30,
	0x10, 0xc5, 0x21, 0x61, 0x37, 0xb3, 0xc0, 0x22, 0x8b, 0x65, 0x2d, 0x4e, 0x51, 0x4e, 0x39, 0xb1,
	0xd2, 0xf6, 0x52, 0xa9, 0xa7, 0x1e, 0xab, 0x8a, 0xaa, 0x32, 0xe2, 0x03, 0x02, 0x71, 0x83, 0x55,
	0x12, 0xa7, 0xb6, 0x55, 0xd1, 0xef, 0xea, 0x4f, 0x54, 0xea, 0x4f, 0x55, 0xe3, 0xa4, 0x10, 0x2a,
	0x0e, 0xdc, 0x7a, 0xf2, 0xbc, 0xa7, 0x19, 0xfb, 0xbd, 0xe7, 0x81, 0x5f, 0xb2, 0xf8, 0x27, 0x8b,
	0x59, 0xa5, 0x95, 0x55, 0x34, 0x70, 0x47, 0xfc, 0x4e, 0xe0, 0xf7, 0xd2, 0x08, 0xcd, 0x45, 0x2e,
	0x8d, 0xc5, 0xf3, 0x89, 0x0e, 0xc1, 0x93, 0x19, 0x23, 0x11, 0x49, 0x7c, 0xee, 0xc9, 0x8c, 0x52,
	0xf0, 0xcb, 0xb4, 0x10, 0xcc, 0x8b, 0x48, 0x12, 0x72, 0x57, 0xd3, 0x31, 0xe0, 0x05, 0x95, 0x61,
	0xdd, 0x88, 0x24, 0x7d, 0x5e, 0x03, 0xec, 0x94, 0x6b, 0x55, 0x32, 0xbf, 0xee, 0xc4, 0x1a, 0x3b,
	0x45, 0x91, 0xca, 0x2d, 0x0b, 0x1c, 0x59, 0x03, 0x64, 0x57, 0x52, 0xdb, 0x0d, 0xeb, 0xd5, 0xac,
	0x03, 0x74, 0x02, 0xbd, 0x42, 0xad, 0xe4, 0x56, 0xb0, 0x1f, 0x8e, 0x6e, 0x10, 0xf2, 0xb9, 0x28,
	0x33, 0xa1, 0xd9, 0xcf, 0x88, 0x24, 0x01, 0x6f, 0x10, 0x2a, 0x15, 0x3b, 0x16, 0x3a, 0x09, 0x9e,
	0xd8, 0xc5, 0xf7, 0x30, 0x3a, 0x36, 0x63, 0x2a, 0xf7, 0xbe, 0xd6, 0x77, 0xca, 0x19, 0x0a, 0x78,
	0x0d, 0xe8, 0x08, 0xba, 0x85, 0xc9, 0x1b, 0x4b, 0x58, 0x62, 0x9f, 0x55, 0x8f, 0xa2, 0x74, 0x8e,
	0x42, 0x5e, 0x83, 0xf8, 0x95, 0xc0, 0x00, 0xaf, 0x5c, 0x56, 0x59, 0x6a, 0xc5, 0xb9, 0xe9, 0x7c,
	0x77, 0x0e, 0x97, 0x30, 0x6c, 0x8b, 0x3e, 0x3f, 0x85, 0xf8, 0x16, 0x06, 0x0b, 0x51, 0x66, 0x8b,
	0x17, 0x33, 0x37, 0x39, 0xda, 0xa5, 0xe0, 0x3f, 0x68, 0x55, 0x34, 0x86, 0x5d, 0x8d, 0xcf, 0x59,
	0xe5, 0xa6, 0x7c, 0xee, 0x59, 0x85, 0xb2, 0x52, 0x6b, 0xd3, 0xf5, 0xa6, 0xd9, 0x86, 0x06, 0xa1,
	0x8c, 0xf6, 0x65, 0xe7, 0xcb, 0xf8, 0xff, 0x46, 0x20, 0xbc, 0x99, 0x2f, 0x84, 0x7e, 0x96, 0x6b,
	0x41, 0xaf, 0xa1, 0xdf, 0xfe, 0x56, 0x3a, 0xa9, 0x77, 0x78, 0xf6, 0x65, 0x71, 0xa7, 0x7f, 0x4f,
	0xf2, 0xa6, 0x8a, 0x3b, 0xf4, 0x0a, 0xe0, 0x90, 0x08, 0x1d, 0xb7, 0x1a, 0xf7, 0x3f, 0x3b, 0xfd,
	0x73, 0x82, 0xfd, 0x1c, 0x3e, 0xf8, 0xd8, 0x0f, 0x1f, 0xe5, 0xb4, 0x1f, 0x3e, 0x36, 0x1c, 0x77,
	0x56, 0x3d, 0xc7, 0x5f, 0x7c, 0x04, 0x00, 0x00, 0xff, 0xff, 0xaa, 0xcf, 0xdf, 0xf6, 0x7e, 0x03,
	0x00, 0x00,
}