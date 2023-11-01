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
	TrustingPeriod       time.Duration `protobuf:"bytes,1,opt,name=trusting_period,json=trustingPeriod,proto3,stdduration" json:"trusting_period"`
	MaxClockDrift        time.Duration `protobuf:"bytes,2,opt,name=max_clock_drift,json=maxClockDrift,proto3,stdduration" json:"max_clock_drift"`
	RefreshThresholdRate *Fraction     `protobuf:"bytes,5,opt,name=refresh_threshold_rate,json=refreshThresholdRate,proto3" json:"refresh_threshold_rate,omitempty"`
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

func (m *ProverConfig) GetRefreshThresholdRate() *Fraction {
	if m != nil {
		return m.RefreshThresholdRate
	}
	return nil
}

type Fraction struct {
	Numerator   uint64 `protobuf:"varint,1,opt,name=numerator,proto3" json:"numerator,omitempty"`
	Denominator uint64 `protobuf:"varint,2,opt,name=denominator,proto3" json:"denominator,omitempty"`
}

func (m *Fraction) Reset()         { *m = Fraction{} }
func (m *Fraction) String() string { return proto.CompactTextString(m) }
func (*Fraction) ProtoMessage()    {}
func (*Fraction) Descriptor() ([]byte, []int) {
	return fileDescriptor_4d00ceb9ab8b08a6, []int{1}
}
func (m *Fraction) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Fraction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Fraction.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Fraction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Fraction.Merge(m, src)
}
func (m *Fraction) XXX_Size() int {
	return m.Size()
}
func (m *Fraction) XXX_DiscardUnknown() {
	xxx_messageInfo_Fraction.DiscardUnknown(m)
}

var xxx_messageInfo_Fraction proto.InternalMessageInfo

func (m *Fraction) GetNumerator() uint64 {
	if m != nil {
		return m.Numerator
	}
	return 0
}

func (m *Fraction) GetDenominator() uint64 {
	if m != nil {
		return m.Denominator
	}
	return 0
}

func init() {
	proto.RegisterType((*ProverConfig)(nil), "relayer.provers.parlia.config.ProverConfig")
	proto.RegisterType((*Fraction)(nil), "relayer.provers.parlia.config.Fraction")
}

func init() {
	proto.RegisterFile("relayer/provers/parlia/config/config.proto", fileDescriptor_4d00ceb9ab8b08a6)
}

var fileDescriptor_4d00ceb9ab8b08a6 = []byte{
	// 367 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xbf, 0x6a, 0xe3, 0x40,
	0x10, 0xc6, 0x25, 0x73, 0x77, 0xf8, 0xd6, 0x77, 0x67, 0x10, 0xe6, 0xf0, 0x99, 0x8b, 0x6c, 0xdc,
	0x24, 0x04, 0xbc, 0x0b, 0xc9, 0x1b, 0xd8, 0x26, 0x90, 0x3f, 0x85, 0x11, 0xa9, 0x02, 0x41, 0xac,
	0xa4, 0x95, 0xb4, 0x44, 0xd2, 0x88, 0xf1, 0x2a, 0x38, 0x6f, 0x90, 0x32, 0x65, 0x1e, 0xc9, 0xa5,
	0xcb, 0x54, 0x49, 0xb0, 0x5f, 0x24, 0x68, 0x25, 0x91, 0x54, 0x21, 0xd5, 0x0e, 0x33, 0xdf, 0xf7,
	0xfb, 0x76, 0x86, 0x1c, 0xa2, 0x48, 0xf8, 0x9d, 0x40, 0x96, 0x23, 0xdc, 0x0a, 0x5c, 0xb2, 0x9c,
	0x63, 0x22, 0x39, 0xf3, 0x21, 0x0b, 0x65, 0x54, 0x3f, 0x34, 0x47, 0x50, 0x60, 0xed, 0xd5, 0x5a,
	0x5a, 0x6b, 0x69, 0xa5, 0xa5, 0x95, 0x68, 0x60, 0x47, 0x00, 0x51, 0x22, 0x98, 0x16, 0x7b, 0x45,
	0xc8, 0x82, 0x02, 0xb9, 0x92, 0x90, 0x55, 0xf6, 0x41, 0x2f, 0x82, 0x08, 0x74, 0xc9, 0xca, 0xaa,
	0xea, 0x8e, 0xef, 0x5b, 0xe4, 0xd7, 0x42, 0xf3, 0x66, 0x1a, 0x63, 0x5d, 0x90, 0xae, 0xc2, 0x62,
	0xa9, 0x64, 0x16, 0xb9, 0xb9, 0x40, 0x09, 0x41, 0xdf, 0x1c, 0x99, 0x07, 0x9d, 0xa3, 0x7f, 0xb4,
	0x0a, 0xa0, 0x4d, 0x00, 0x9d, 0xd7, 0x01, 0xd3, 0xf6, 0xfa, 0x79, 0x68, 0x3c, 0xbe, 0x0c, 0x4d,
	0xe7, 0x4f, 0xe3, 0x5d, 0x68, 0xab, 0x75, 0x4e, 0xba, 0x29, 0x5f, 0xb9, 0x7e, 0x02, 0xfe, 0x8d,
	0x1b, 0xa0, 0x0c, 0x55, 0xbf, 0xf5, 0x75, 0xda, 0xef, 0x94, 0xaf, 0x66, 0xa5, 0x75, 0x5e, 0x3a,
	0xad, 0x6b, 0xf2, 0x17, 0x45, 0x88, 0x62, 0x19, 0xbb, 0x2a, 0x2e, 0x1f, 0x48, 0x02, 0x17, 0xb9,
	0x12, 0xfd, 0xef, 0x9a, 0xb9, 0x4f, 0x3f, 0xbd, 0x10, 0x3d, 0x41, 0xee, 0x97, 0x09, 0x4e, 0xaf,
	0xc6, 0x5c, 0x36, 0x14, 0x87, 0x2b, 0x31, 0x3e, 0x23, 0xed, 0x46, 0x61, 0xfd, 0x27, 0x3f, 0xb3,
	0x22, 0x15, 0xc8, 0x15, 0xa0, 0xde, 0xff, 0x9b, 0xf3, 0xde, 0xb0, 0x46, 0xa4, 0x13, 0x88, 0x0c,
	0x52, 0x99, 0xe9, 0x79, 0x4b, 0xcf, 0x3f, 0xb6, 0xa6, 0xa7, 0xeb, 0xad, 0x6d, 0x6e, 0xb6, 0xb6,
	0xf9, 0xba, 0xb5, 0xcd, 0x87, 0x9d, 0x6d, 0x6c, 0x76, 0xb6, 0xf1, 0xb4, 0xb3, 0x8d, 0x2b, 0x16,
	0x49, 0x15, 0x17, 0x1e, 0xf5, 0x21, 0x65, 0x01, 0x57, 0xdc, 0x8f, 0xb9, 0xcc, 0x12, 0xee, 0x31,
	0xe9, 0xf9, 0x93, 0xea, 0xbf, 0x13, 0xbd, 0x06, 0x4b, 0x21, 0x28, 0x12, 0xe1, 0xfd, 0xd0, 0x17,
	0x3a, 0x7e, 0x0b, 0x00, 0x00, 0xff, 0xff, 0xd0, 0x5c, 0x69, 0x66, 0x2b, 0x02, 0x00, 0x00,
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
	if m.RefreshThresholdRate != nil {
		{
			size, err := m.RefreshThresholdRate.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintConfig(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	n2, err2 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.MaxClockDrift, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MaxClockDrift):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintConfig(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	n3, err3 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.TrustingPeriod, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.TrustingPeriod):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintConfig(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Fraction) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Fraction) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Fraction) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Denominator != 0 {
		i = encodeVarintConfig(dAtA, i, uint64(m.Denominator))
		i--
		dAtA[i] = 0x10
	}
	if m.Numerator != 0 {
		i = encodeVarintConfig(dAtA, i, uint64(m.Numerator))
		i--
		dAtA[i] = 0x8
	}
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
	if m.RefreshThresholdRate != nil {
		l = m.RefreshThresholdRate.Size()
		n += 1 + l + sovConfig(uint64(l))
	}
	return n
}

func (m *Fraction) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Numerator != 0 {
		n += 1 + sovConfig(uint64(m.Numerator))
	}
	if m.Denominator != 0 {
		n += 1 + sovConfig(uint64(m.Denominator))
	}
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
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefreshThresholdRate", wireType)
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
			if m.RefreshThresholdRate == nil {
				m.RefreshThresholdRate = &Fraction{}
			}
			if err := m.RefreshThresholdRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *Fraction) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: Fraction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Fraction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Numerator", wireType)
			}
			m.Numerator = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Numerator |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denominator", wireType)
			}
			m.Denominator = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Denominator |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
