// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: relayer/provers/parlia/config/config.proto

package module

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ProverConfig struct {
	TrustingPeriod time.Duration `protobuf:"bytes,1,opt,name=trusting_period,json=trustingPeriod,proto3,stdduration" json:"trusting_period"`
	MaxClockDrift  time.Duration `protobuf:"bytes,2,opt,name=max_clock_drift,json=maxClockDrift,proto3,stdduration" json:"max_clock_drift"`
}

func (m *ProverConfig) Reset()         { *m = ProverConfig{} }
func (m *ProverConfig) String() string { return proto.CompactTextString(m) }
func (*ProverConfig) ProtoMessage()    {}
func (*ProverConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_4d00ceb9ab8b08a6, []int{0}
}
func (m *ProverConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProverConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProverConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ProverConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProverConfig.Merge(m, src)
}
func (m *ProverConfig) XXX_Size() int {
	return m.Size()
}
func (m *ProverConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ProverConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ProverConfig proto.InternalMessageInfo

func (m *ProverConfig) GetTrustingPeriod() time.Duration {
	if m != nil {
		return m.TrustingPeriod
	}
	return 0
}

func (m *ProverConfig) GetMaxClockDrift() time.Duration {
	if m != nil {
		return m.MaxClockDrift
	}
	return 0
}

func init() {
	proto.RegisterType((*ProverConfig)(nil), "relayer.provers.parlia.config.ProverConfig")
}

func init() {
	proto.RegisterFile("relayer/provers/parlia/config/config.proto", fileDescriptor_4d00ceb9ab8b08a6)
}

var fileDescriptor_4d00ceb9ab8b08a6 = []byte{
	// 283 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x3f, 0x4e, 0xc3, 0x30,
	0x18, 0xc5, 0x63, 0x06, 0x84, 0xc2, 0x9f, 0x4a, 0x11, 0x43, 0xa9, 0x84, 0x8b, 0x98, 0x10, 0x52,
	0x6d, 0x09, 0x6e, 0xd0, 0x76, 0x41, 0x30, 0x54, 0x8c, 0x2c, 0x91, 0xe3, 0x38, 0xae, 0x85, 0x93,
	0x2f, 0x72, 0x6d, 0x54, 0x6e, 0xc1, 0xc8, 0x15, 0xb8, 0x49, 0xc7, 0x8e, 0x4c, 0x80, 0x92, 0x8b,
	0xa0, 0xd8, 0xe9, 0xde, 0xc9, 0x9f, 0xac, 0xdf, 0xef, 0x3d, 0xe9, 0xc5, 0xb7, 0x46, 0x68, 0xf6,
	0x2e, 0x0c, 0xad, 0x0d, 0xbc, 0x09, 0xb3, 0xa2, 0x35, 0x33, 0x5a, 0x31, 0xca, 0xa1, 0x2a, 0x94,
	0xec, 0x1f, 0x52, 0x1b, 0xb0, 0x90, 0x5c, 0xf6, 0x2c, 0xe9, 0x59, 0x12, 0x58, 0x12, 0xa0, 0x11,
	0x96, 0x00, 0x52, 0x0b, 0xea, 0xe1, 0xcc, 0x15, 0x34, 0x77, 0x86, 0x59, 0x05, 0x55, 0xd0, 0x47,
	0xe7, 0x12, 0x24, 0xf8, 0x93, 0x76, 0x57, 0xf8, 0xbd, 0xfe, 0x42, 0xf1, 0xc9, 0xc2, 0xe7, 0xcd,
	0x7c, 0x4c, 0xf2, 0x14, 0x0f, 0xac, 0x71, 0x2b, 0xab, 0x2a, 0x99, 0xd6, 0xc2, 0x28, 0xc8, 0x87,
	0xe8, 0x0a, 0xdd, 0x1c, 0xdf, 0x5d, 0x90, 0x50, 0x40, 0x76, 0x05, 0x64, 0xde, 0x17, 0x4c, 0x8f,
	0x36, 0x3f, 0xe3, 0xe8, 0xf3, 0x77, 0x8c, 0x9e, 0xcf, 0x76, 0xee, 0xc2, 0xab, 0xc9, 0x63, 0x3c,
	0x28, 0xd9, 0x3a, 0xe5, 0x1a, 0xf8, 0x6b, 0x9a, 0x1b, 0x55, 0xd8, 0xe1, 0xc1, 0xfe, 0x69, 0xa7,
	0x25, 0x5b, 0xcf, 0x3a, 0x75, 0xde, 0x99, 0xd3, 0x87, 0x4d, 0x83, 0xd1, 0xb6, 0xc1, 0xe8, 0xaf,
	0xc1, 0xe8, 0xa3, 0xc5, 0xd1, 0xb6, 0xc5, 0xd1, 0x77, 0x8b, 0xa3, 0x17, 0x2a, 0x95, 0x5d, 0xba,
	0x8c, 0x70, 0x28, 0x69, 0xce, 0x2c, 0xe3, 0x4b, 0xa6, 0x2a, 0xcd, 0x32, 0xaa, 0x32, 0x3e, 0x09,
	0x33, 0x4d, 0xfc, 0x7a, 0xb4, 0x84, 0xdc, 0x69, 0x91, 0x1d, 0xfa, 0xda, 0xfb, 0xff, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x5a, 0x4f, 0xa4, 0xee, 0x80, 0x01, 0x00, 0x00,
}

func (m *ProverConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProverConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProverConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.MaxClockDrift, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MaxClockDrift):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintConfig(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x12
	n2, err2 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.TrustingPeriod, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.TrustingPeriod):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintConfig(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintConfig(dAtA []byte, offset int, v uint64) int {
	offset -= sovConfig(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ProverConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.TrustingPeriod)
	n += 1 + l + sovConfig(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MaxClockDrift)
	n += 1 + l + sovConfig(uint64(l))
	return n
}

func sovConfig(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozConfig(x uint64) (n int) {
	return sovConfig(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ProverConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProverConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProverConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrustingPeriod", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.TrustingPeriod, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxClockDrift", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.MaxClockDrift, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthConfig
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
func skipConfig(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConfig
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
					return 0, ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowConfig
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
			if length < 0 {
				return 0, ErrInvalidLengthConfig
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupConfig
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthConfig
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthConfig        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConfig          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupConfig = fmt.Errorf("proto: unexpected end of group")
)
