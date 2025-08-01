// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: babylon/finality/v1/events.proto

package types

import (
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// EventSlashedFinalityProvider is the event emitted when a finality provider is slashed
// due to signing two conflicting blocks
type EventSlashedFinalityProvider struct {
	// evidence is the evidence that the finality provider double signs
	Evidence *Evidence `protobuf:"bytes,1,opt,name=evidence,proto3" json:"evidence,omitempty"`
}

func (m *EventSlashedFinalityProvider) Reset()         { *m = EventSlashedFinalityProvider{} }
func (m *EventSlashedFinalityProvider) String() string { return proto.CompactTextString(m) }
func (*EventSlashedFinalityProvider) ProtoMessage()    {}
func (*EventSlashedFinalityProvider) Descriptor() ([]byte, []int) {
	return fileDescriptor_c34c03aae5e3e6bf, []int{0}
}
func (m *EventSlashedFinalityProvider) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventSlashedFinalityProvider) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventSlashedFinalityProvider.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventSlashedFinalityProvider) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventSlashedFinalityProvider.Merge(m, src)
}
func (m *EventSlashedFinalityProvider) XXX_Size() int {
	return m.Size()
}
func (m *EventSlashedFinalityProvider) XXX_DiscardUnknown() {
	xxx_messageInfo_EventSlashedFinalityProvider.DiscardUnknown(m)
}

var xxx_messageInfo_EventSlashedFinalityProvider proto.InternalMessageInfo

func (m *EventSlashedFinalityProvider) GetEvidence() *Evidence {
	if m != nil {
		return m.Evidence
	}
	return nil
}

// EventJailedFinalityProvider is the event emitted when a finality provider is
// jailed due to inactivity
type EventJailedFinalityProvider struct {
	// public_key is the BTC public key of the finality provider
	PublicKey string `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (m *EventJailedFinalityProvider) Reset()         { *m = EventJailedFinalityProvider{} }
func (m *EventJailedFinalityProvider) String() string { return proto.CompactTextString(m) }
func (*EventJailedFinalityProvider) ProtoMessage()    {}
func (*EventJailedFinalityProvider) Descriptor() ([]byte, []int) {
	return fileDescriptor_c34c03aae5e3e6bf, []int{1}
}
func (m *EventJailedFinalityProvider) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventJailedFinalityProvider) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventJailedFinalityProvider.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventJailedFinalityProvider) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventJailedFinalityProvider.Merge(m, src)
}
func (m *EventJailedFinalityProvider) XXX_Size() int {
	return m.Size()
}
func (m *EventJailedFinalityProvider) XXX_DiscardUnknown() {
	xxx_messageInfo_EventJailedFinalityProvider.DiscardUnknown(m)
}

var xxx_messageInfo_EventJailedFinalityProvider proto.InternalMessageInfo

func (m *EventJailedFinalityProvider) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func init() {
	proto.RegisterType((*EventSlashedFinalityProvider)(nil), "babylon.finality.v1.EventSlashedFinalityProvider")
	proto.RegisterType((*EventJailedFinalityProvider)(nil), "babylon.finality.v1.EventJailedFinalityProvider")
}

func init() { proto.RegisterFile("babylon/finality/v1/events.proto", fileDescriptor_c34c03aae5e3e6bf) }

var fileDescriptor_c34c03aae5e3e6bf = []byte{
	// 234 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0x4a, 0x4c, 0xaa,
	0xcc, 0xc9, 0xcf, 0xd3, 0x4f, 0xcb, 0xcc, 0x4b, 0xcc, 0xc9, 0x2c, 0xa9, 0xd4, 0x2f, 0x33, 0xd4,
	0x4f, 0x2d, 0x4b, 0xcd, 0x2b, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x86, 0xaa,
	0xd0, 0x83, 0xa9, 0xd0, 0x2b, 0x33, 0x94, 0x52, 0xc2, 0xa6, 0x0d, 0xae, 0x00, 0xac, 0x51, 0x29,
	0x92, 0x4b, 0xc6, 0x15, 0x64, 0x50, 0x70, 0x4e, 0x62, 0x71, 0x46, 0x6a, 0x8a, 0x1b, 0x54, 0x36,
	0xa0, 0x28, 0xbf, 0x2c, 0x33, 0x25, 0xb5, 0x48, 0xc8, 0x92, 0x8b, 0x23, 0x15, 0xc4, 0xca, 0x4b,
	0x4e, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x36, 0x92, 0xd5, 0xc3, 0x62, 0x97, 0x9e, 0x2b, 0x54,
	0x51, 0x10, 0x5c, 0xb9, 0x92, 0x0d, 0x97, 0x34, 0xd8, 0x68, 0xaf, 0xc4, 0xcc, 0x1c, 0x2c, 0x26,
	0xcb, 0x72, 0x71, 0x15, 0x94, 0x26, 0xe5, 0x64, 0x26, 0xc7, 0x67, 0xa7, 0x56, 0x82, 0xcd, 0xe6,
	0x0c, 0xe2, 0x84, 0x88, 0x78, 0xa7, 0x56, 0x3a, 0xf9, 0x9f, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91,
	0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3,
	0xb1, 0x1c, 0x43, 0x94, 0x69, 0x7a, 0x66, 0x49, 0x46, 0x69, 0x92, 0x5e, 0x72, 0x7e, 0xae, 0x3e,
	0xd4, 0x29, 0x39, 0x89, 0x49, 0xc5, 0xba, 0x99, 0xf9, 0x30, 0xae, 0x7e, 0x99, 0xb1, 0x7e, 0x05,
	0xc2, 0xd7, 0x25, 0x95, 0x05, 0xa9, 0xc5, 0x49, 0x6c, 0x60, 0x0f, 0x1b, 0x03, 0x02, 0x00, 0x00,
	0xff, 0xff, 0x7a, 0x09, 0x84, 0xea, 0x4d, 0x01, 0x00, 0x00,
}

func (m *EventSlashedFinalityProvider) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventSlashedFinalityProvider) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventSlashedFinalityProvider) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Evidence != nil {
		{
			size, err := m.Evidence.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintEvents(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventJailedFinalityProvider) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventJailedFinalityProvider) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventJailedFinalityProvider) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PublicKey) > 0 {
		i -= len(m.PublicKey)
		copy(dAtA[i:], m.PublicKey)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.PublicKey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintEvents(dAtA []byte, offset int, v uint64) int {
	offset -= sovEvents(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EventSlashedFinalityProvider) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Evidence != nil {
		l = m.Evidence.Size()
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func (m *EventJailedFinalityProvider) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func sovEvents(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvents(x uint64) (n int) {
	return sovEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventSlashedFinalityProvider) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventSlashedFinalityProvider: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventSlashedFinalityProvider: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Evidence", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Evidence == nil {
				m.Evidence = &Evidence{}
			}
			if err := m.Evidence.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func (m *EventJailedFinalityProvider) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventJailedFinalityProvider: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventJailedFinalityProvider: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func skipEvents(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
				return 0, ErrInvalidLengthEvents
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEvents
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEvents
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEvents        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvents          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEvents = fmt.Errorf("proto: unexpected end of group")
)
