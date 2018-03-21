// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: envoy/config/filter/http/buffer/v2/buffer.proto

/*
	Package v2 is a generated protocol buffer package.

	It is generated from these files:
		envoy/config/filter/http/buffer/v2/buffer.proto

	It has these top-level messages:
		Buffer
*/
package v2

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/types"
import google_protobuf1 "github.com/gogo/protobuf/types"
import _ "github.com/lyft/protoc-gen-validate/validate"
import _ "github.com/gogo/protobuf/gogoproto"

import time "time"

import github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Buffer struct {
	// The maximum request size that the filter will buffer before the connection
	// manager will stop buffering and return a 413 response.
	MaxRequestBytes *google_protobuf1.UInt32Value `protobuf:"bytes,1,opt,name=max_request_bytes,json=maxRequestBytes" json:"max_request_bytes,omitempty"`
	// The maximum number of seconds that the filter will wait for a complete
	// request before returning a 408 response.
	MaxRequestTime *time.Duration `protobuf:"bytes,2,opt,name=max_request_time,json=maxRequestTime,stdduration" json:"max_request_time,omitempty"`
}

func (m *Buffer) Reset()                    { *m = Buffer{} }
func (m *Buffer) String() string            { return proto.CompactTextString(m) }
func (*Buffer) ProtoMessage()               {}
func (*Buffer) Descriptor() ([]byte, []int) { return fileDescriptorBuffer, []int{0} }

func (m *Buffer) GetMaxRequestBytes() *google_protobuf1.UInt32Value {
	if m != nil {
		return m.MaxRequestBytes
	}
	return nil
}

func (m *Buffer) GetMaxRequestTime() *time.Duration {
	if m != nil {
		return m.MaxRequestTime
	}
	return nil
}

func init() {
	proto.RegisterType((*Buffer)(nil), "envoy.config.filter.http.buffer.v2.Buffer")
}
func (m *Buffer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Buffer) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.MaxRequestBytes != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintBuffer(dAtA, i, uint64(m.MaxRequestBytes.Size()))
		n1, err := m.MaxRequestBytes.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.MaxRequestTime != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintBuffer(dAtA, i, uint64(github_com_gogo_protobuf_types.SizeOfStdDuration(*m.MaxRequestTime)))
		n2, err := github_com_gogo_protobuf_types.StdDurationMarshalTo(*m.MaxRequestTime, dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func encodeVarintBuffer(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Buffer) Size() (n int) {
	var l int
	_ = l
	if m.MaxRequestBytes != nil {
		l = m.MaxRequestBytes.Size()
		n += 1 + l + sovBuffer(uint64(l))
	}
	if m.MaxRequestTime != nil {
		l = github_com_gogo_protobuf_types.SizeOfStdDuration(*m.MaxRequestTime)
		n += 1 + l + sovBuffer(uint64(l))
	}
	return n
}

func sovBuffer(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozBuffer(x uint64) (n int) {
	return sovBuffer(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Buffer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBuffer
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Buffer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Buffer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRequestBytes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBuffer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthBuffer
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.MaxRequestBytes == nil {
				m.MaxRequestBytes = &google_protobuf1.UInt32Value{}
			}
			if err := m.MaxRequestBytes.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRequestTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBuffer
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthBuffer
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.MaxRequestTime == nil {
				m.MaxRequestTime = new(time.Duration)
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(m.MaxRequestTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBuffer(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBuffer
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipBuffer(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBuffer
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowBuffer
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowBuffer
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthBuffer
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowBuffer
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipBuffer(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthBuffer = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBuffer   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("envoy/config/filter/http/buffer/v2/buffer.proto", fileDescriptorBuffer)
}

var fileDescriptorBuffer = []byte{
	// 293 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x8f, 0x31, 0x4e, 0xc3, 0x30,
	0x14, 0x86, 0xeb, 0x50, 0x15, 0x08, 0x12, 0x84, 0x08, 0x89, 0x82, 0x50, 0xa8, 0x3a, 0xa1, 0x0e,
	0xb6, 0x94, 0xde, 0x20, 0x62, 0x61, 0x0d, 0x94, 0x81, 0xa5, 0x72, 0xe8, 0x4b, 0xb0, 0x94, 0xc4,
	0xc1, 0x79, 0x09, 0xed, 0x4d, 0x38, 0x03, 0x33, 0x13, 0x13, 0x23, 0x23, 0x37, 0x00, 0x65, 0xe3,
	0x16, 0x28, 0x76, 0x2a, 0x90, 0xba, 0xfd, 0xf2, 0xef, 0xef, 0xfb, 0xf5, 0x6c, 0x06, 0x79, 0x2d,
	0x57, 0xec, 0x5e, 0xe6, 0xb1, 0x48, 0x58, 0x2c, 0x52, 0x04, 0xc5, 0x1e, 0x10, 0x0b, 0x16, 0x55,
	0x71, 0x0c, 0x8a, 0xd5, 0x7e, 0x97, 0x68, 0xa1, 0x24, 0x4a, 0x77, 0xac, 0x01, 0x6a, 0x00, 0x6a,
	0x00, 0xda, 0x02, 0xb4, 0xfb, 0x56, 0xfb, 0xa7, 0x5e, 0x22, 0x65, 0x92, 0x02, 0xd3, 0x44, 0x54,
	0xc5, 0x6c, 0x51, 0x29, 0x8e, 0x42, 0xe6, 0xc6, 0xb1, 0xd9, 0x3f, 0x29, 0x5e, 0x14, 0xa0, 0xca,
	0xae, 0x3f, 0xae, 0x79, 0x2a, 0x16, 0x1c, 0x81, 0xad, 0x43, 0x57, 0x1c, 0x25, 0x32, 0x91, 0x3a,
	0xb2, 0x36, 0x99, 0xd7, 0xf1, 0x2b, 0xb1, 0x07, 0x81, 0x1e, 0x77, 0xaf, 0xed, 0xc3, 0x8c, 0x2f,
	0xe7, 0x0a, 0x1e, 0x2b, 0x28, 0x71, 0x1e, 0xad, 0x10, 0xca, 0x21, 0x19, 0x91, 0x8b, 0x3d, 0xff,
	0x8c, 0x9a, 0x55, 0xba, 0x5e, 0xa5, 0xb3, 0xab, 0x1c, 0xa7, 0xfe, 0x2d, 0x4f, 0x2b, 0x08, 0x76,
	0xdf, 0x7e, 0xde, 0xb7, 0xfa, 0x13, 0x6b, 0xd4, 0x0b, 0x0f, 0x32, 0xbe, 0x0c, 0x8d, 0x20, 0x68,
	0x79, 0x77, 0x66, 0x3b, 0xff, 0xa5, 0x28, 0x32, 0x18, 0x5a, 0xda, 0x79, 0xb2, 0xe1, 0xbc, 0xec,
	0x2e, 0x0d, 0x9c, 0xe7, 0xaf, 0x73, 0xd2, 0x4a, 0xb7, 0x5f, 0x48, 0x7f, 0x87, 0x4c, 0x7a, 0xe1,
	0xfe, 0x9f, 0xf7, 0x46, 0x64, 0x10, 0x38, 0x1f, 0x8d, 0x47, 0x3e, 0x1b, 0x8f, 0x7c, 0x37, 0x1e,
	0xb9, 0xb3, 0x6a, 0x3f, 0x1a, 0x68, 0xcd, 0xf4, 0x37, 0x00, 0x00, 0xff, 0xff, 0xbe, 0x3f, 0x43,
	0xb3, 0x95, 0x01, 0x00, 0x00,
}
