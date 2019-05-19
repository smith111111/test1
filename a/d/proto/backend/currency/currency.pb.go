// Code generated by protoc-gen-go. DO NOT EDIT.
// source: backend/currency/currency.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	backend/currency/currency.proto

It has these top-level messages:
	GetCurrencyReq
	CryptoCurrencyInfo
	FiatCurrencyInfo
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

type GetCurrencyReq struct {
	Code string `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
}

func (m *GetCurrencyReq) Reset()                    { *m = GetCurrencyReq{} }
func (m *GetCurrencyReq) String() string            { return proto1.CompactTextString(m) }
func (*GetCurrencyReq) ProtoMessage()               {}
func (*GetCurrencyReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *GetCurrencyReq) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

type CryptoCurrencyInfo struct {
}

func (m *CryptoCurrencyInfo) Reset()                    { *m = CryptoCurrencyInfo{} }
func (m *CryptoCurrencyInfo) String() string            { return proto1.CompactTextString(m) }
func (*CryptoCurrencyInfo) ProtoMessage()               {}
func (*CryptoCurrencyInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type FiatCurrencyInfo struct {
}

func (m *FiatCurrencyInfo) Reset()                    { *m = FiatCurrencyInfo{} }
func (m *FiatCurrencyInfo) String() string            { return proto1.CompactTextString(m) }
func (*FiatCurrencyInfo) ProtoMessage()               {}
func (*FiatCurrencyInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func init() {
	proto1.RegisterType((*GetCurrencyReq)(nil), "proto.GetCurrencyReq")
	proto1.RegisterType((*CryptoCurrencyInfo)(nil), "proto.CryptoCurrencyInfo")
	proto1.RegisterType((*FiatCurrencyInfo)(nil), "proto.FiatCurrencyInfo")
}

func init() { proto1.RegisterFile("backend/currency/currency.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 170 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4f, 0x4a, 0x4c, 0xce,
	0x4e, 0xcd, 0x4b, 0xd1, 0x4f, 0x2e, 0x2d, 0x2a, 0x4a, 0xcd, 0x4b, 0xae, 0x84, 0x33, 0xf4, 0x0a,
	0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x58, 0xc1, 0x94, 0x92, 0x0a, 0x17, 0x9f, 0x7b, 0x6a, 0x89, 0x33,
	0x54, 0x2e, 0x28, 0xb5, 0x50, 0x48, 0x88, 0x8b, 0x25, 0x39, 0x3f, 0x25, 0x55, 0x82, 0x51, 0x81,
	0x51, 0x83, 0x33, 0x08, 0xcc, 0x56, 0x12, 0xe1, 0x12, 0x72, 0x2e, 0xaa, 0x2c, 0x28, 0xc9, 0x87,
	0x29, 0xf4, 0xcc, 0x4b, 0xcb, 0x57, 0x12, 0xe2, 0x12, 0x70, 0xcb, 0x4c, 0x2c, 0x41, 0x16, 0x33,
	0x9a, 0xcf, 0xc8, 0xc5, 0x0f, 0x13, 0x08, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0x15, 0x72, 0xe7,
	0x12, 0x04, 0xd9, 0x81, 0x62, 0x80, 0x90, 0x28, 0xc4, 0x1d, 0x7a, 0xa8, 0xb6, 0x4b, 0x49, 0x42,
	0x85, 0xb1, 0x58, 0xc7, 0x20, 0xe4, 0xcc, 0xc5, 0xef, 0x9e, 0x5a, 0x82, 0x6c, 0x27, 0x2e, 0x63,
	0xc4, 0xa1, 0xc2, 0xe8, 0xee, 0x53, 0x62, 0x48, 0x62, 0x03, 0xcb, 0x18, 0x03, 0x02, 0x00, 0x00,
	0xff, 0xff, 0x6b, 0xd0, 0xb1, 0x35, 0x22, 0x01, 0x00, 0x00,
}