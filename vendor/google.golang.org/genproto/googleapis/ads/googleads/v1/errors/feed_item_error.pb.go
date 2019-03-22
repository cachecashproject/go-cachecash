// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/errors/feed_item_error.proto

package errors // import "google.golang.org/genproto/googleapis/ads/googleads/v1/errors"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Enum describing possible feed item errors.
type FeedItemErrorEnum_FeedItemError int32

const (
	// Enum unspecified.
	FeedItemErrorEnum_UNSPECIFIED FeedItemErrorEnum_FeedItemError = 0
	// The received error code is not known in this version.
	FeedItemErrorEnum_UNKNOWN FeedItemErrorEnum_FeedItemError = 1
	// Cannot convert the feed attribute value from string to its real type.
	FeedItemErrorEnum_CANNOT_CONVERT_ATTRIBUTE_VALUE_FROM_STRING FeedItemErrorEnum_FeedItemError = 2
	// Cannot operate on removed feed item.
	FeedItemErrorEnum_CANNOT_OPERATE_ON_REMOVED_FEED_ITEM FeedItemErrorEnum_FeedItemError = 3
	// Date time zone does not match the account's time zone.
	FeedItemErrorEnum_DATE_TIME_MUST_BE_IN_ACCOUNT_TIME_ZONE FeedItemErrorEnum_FeedItemError = 4
	// Feed item with the key attributes could not be found.
	FeedItemErrorEnum_KEY_ATTRIBUTES_NOT_FOUND FeedItemErrorEnum_FeedItemError = 5
	// Url feed attribute value is not valid.
	FeedItemErrorEnum_INVALID_URL FeedItemErrorEnum_FeedItemError = 6
	// Some key attributes are missing.
	FeedItemErrorEnum_MISSING_KEY_ATTRIBUTES FeedItemErrorEnum_FeedItemError = 7
	// Feed item has same key attributes as another feed item.
	FeedItemErrorEnum_KEY_ATTRIBUTES_NOT_UNIQUE FeedItemErrorEnum_FeedItemError = 8
	// Cannot modify key attributes on an existing feed item.
	FeedItemErrorEnum_CANNOT_MODIFY_KEY_ATTRIBUTE_VALUE FeedItemErrorEnum_FeedItemError = 9
	// The feed attribute value is too large.
	FeedItemErrorEnum_SIZE_TOO_LARGE_FOR_MULTI_VALUE_ATTRIBUTE FeedItemErrorEnum_FeedItemError = 10
)

var FeedItemErrorEnum_FeedItemError_name = map[int32]string{
	0:  "UNSPECIFIED",
	1:  "UNKNOWN",
	2:  "CANNOT_CONVERT_ATTRIBUTE_VALUE_FROM_STRING",
	3:  "CANNOT_OPERATE_ON_REMOVED_FEED_ITEM",
	4:  "DATE_TIME_MUST_BE_IN_ACCOUNT_TIME_ZONE",
	5:  "KEY_ATTRIBUTES_NOT_FOUND",
	6:  "INVALID_URL",
	7:  "MISSING_KEY_ATTRIBUTES",
	8:  "KEY_ATTRIBUTES_NOT_UNIQUE",
	9:  "CANNOT_MODIFY_KEY_ATTRIBUTE_VALUE",
	10: "SIZE_TOO_LARGE_FOR_MULTI_VALUE_ATTRIBUTE",
}
var FeedItemErrorEnum_FeedItemError_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UNKNOWN":     1,
	"CANNOT_CONVERT_ATTRIBUTE_VALUE_FROM_STRING": 2,
	"CANNOT_OPERATE_ON_REMOVED_FEED_ITEM":        3,
	"DATE_TIME_MUST_BE_IN_ACCOUNT_TIME_ZONE":     4,
	"KEY_ATTRIBUTES_NOT_FOUND":                   5,
	"INVALID_URL":                                6,
	"MISSING_KEY_ATTRIBUTES":                     7,
	"KEY_ATTRIBUTES_NOT_UNIQUE":                  8,
	"CANNOT_MODIFY_KEY_ATTRIBUTE_VALUE":          9,
	"SIZE_TOO_LARGE_FOR_MULTI_VALUE_ATTRIBUTE":   10,
}

func (x FeedItemErrorEnum_FeedItemError) String() string {
	return proto.EnumName(FeedItemErrorEnum_FeedItemError_name, int32(x))
}
func (FeedItemErrorEnum_FeedItemError) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_error_88a9937a4741c898, []int{0, 0}
}

// Container for enum describing possible feed item errors.
type FeedItemErrorEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FeedItemErrorEnum) Reset()         { *m = FeedItemErrorEnum{} }
func (m *FeedItemErrorEnum) String() string { return proto.CompactTextString(m) }
func (*FeedItemErrorEnum) ProtoMessage()    {}
func (*FeedItemErrorEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_feed_item_error_88a9937a4741c898, []int{0}
}
func (m *FeedItemErrorEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FeedItemErrorEnum.Unmarshal(m, b)
}
func (m *FeedItemErrorEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FeedItemErrorEnum.Marshal(b, m, deterministic)
}
func (dst *FeedItemErrorEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeedItemErrorEnum.Merge(dst, src)
}
func (m *FeedItemErrorEnum) XXX_Size() int {
	return xxx_messageInfo_FeedItemErrorEnum.Size(m)
}
func (m *FeedItemErrorEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_FeedItemErrorEnum.DiscardUnknown(m)
}

var xxx_messageInfo_FeedItemErrorEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*FeedItemErrorEnum)(nil), "google.ads.googleads.v1.errors.FeedItemErrorEnum")
	proto.RegisterEnum("google.ads.googleads.v1.errors.FeedItemErrorEnum_FeedItemError", FeedItemErrorEnum_FeedItemError_name, FeedItemErrorEnum_FeedItemError_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/errors/feed_item_error.proto", fileDescriptor_feed_item_error_88a9937a4741c898)
}

var fileDescriptor_feed_item_error_88a9937a4741c898 = []byte{
	// 485 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xdd, 0x8a, 0xd3, 0x40,
	0x18, 0x75, 0x5b, 0xdd, 0xd5, 0x59, 0xc4, 0x71, 0x2e, 0x44, 0x97, 0x75, 0xc1, 0x8a, 0x3f, 0x2c,
	0x92, 0x50, 0xf4, 0x2a, 0x5e, 0x4d, 0x9b, 0x2f, 0x65, 0xd8, 0x66, 0xa6, 0x26, 0x33, 0x91, 0x2d,
	0x85, 0x8f, 0x6a, 0x62, 0x28, 0x6c, 0x33, 0xa5, 0xa9, 0xfb, 0x0a, 0xbe, 0x87, 0x97, 0x3e, 0x8a,
	0x8f, 0x22, 0x78, 0xe1, 0x1b, 0x48, 0x3a, 0x6d, 0xa5, 0xa0, 0x5e, 0xe5, 0x70, 0x72, 0xce, 0xf9,
	0x66, 0xe6, 0x7c, 0xe4, 0x4d, 0x69, 0x6d, 0x79, 0x55, 0xf8, 0xd3, 0xbc, 0xf6, 0x1d, 0x6c, 0xd0,
	0x75, 0xd7, 0x2f, 0x96, 0x4b, 0xbb, 0xac, 0xfd, 0x4f, 0x45, 0x91, 0xe3, 0x6c, 0x55, 0xcc, 0x71,
	0x4d, 0x78, 0x8b, 0xa5, 0x5d, 0x59, 0x76, 0xe6, 0xa4, 0xde, 0x34, 0xaf, 0xbd, 0x9d, 0xcb, 0xbb,
	0xee, 0x7a, 0xce, 0x75, 0x72, 0xba, 0x4d, 0x5d, 0xcc, 0xfc, 0x69, 0x55, 0xd9, 0xd5, 0x74, 0x35,
	0xb3, 0x55, 0xed, 0xdc, 0x9d, 0x2f, 0x6d, 0x72, 0x3f, 0x2a, 0x8a, 0x5c, 0xac, 0x8a, 0x39, 0x34,
	0x06, 0xa8, 0x3e, 0xcf, 0x3b, 0xbf, 0x5a, 0xe4, 0xee, 0x1e, 0xcb, 0xee, 0x91, 0x63, 0x23, 0xd3,
	0x11, 0xf4, 0x45, 0x24, 0x20, 0xa4, 0x37, 0xd8, 0x31, 0x39, 0x32, 0xf2, 0x42, 0xaa, 0xf7, 0x92,
	0x1e, 0x30, 0x8f, 0x9c, 0xf7, 0xb9, 0x94, 0x4a, 0x63, 0x5f, 0xc9, 0x0c, 0x12, 0x8d, 0x5c, 0xeb,
	0x44, 0xf4, 0x8c, 0x06, 0xcc, 0xf8, 0xd0, 0x00, 0x46, 0x89, 0x8a, 0x31, 0xd5, 0x89, 0x90, 0x03,
	0xda, 0x62, 0x2f, 0xc8, 0xd3, 0x8d, 0x5e, 0x8d, 0x20, 0xe1, 0x1a, 0x50, 0x49, 0x4c, 0x20, 0x56,
	0x19, 0x84, 0x18, 0x01, 0x84, 0x28, 0x34, 0xc4, 0xb4, 0xcd, 0xce, 0xc9, 0xf3, 0xb0, 0xf9, 0xad,
	0x45, 0x0c, 0x18, 0x9b, 0x54, 0x63, 0x0f, 0x50, 0x48, 0xe4, 0xfd, 0xbe, 0x32, 0x52, 0x3b, 0x7e,
	0xac, 0x24, 0xd0, 0x9b, 0xec, 0x94, 0x3c, 0xbc, 0x80, 0xcb, 0x3f, 0x93, 0x53, 0x6c, 0x06, 0x44,
	0xca, 0xc8, 0x90, 0xde, 0x6a, 0x2e, 0x20, 0x64, 0xc6, 0x87, 0x22, 0x44, 0x93, 0x0c, 0xe9, 0x21,
	0x3b, 0x21, 0x0f, 0x62, 0x91, 0xa6, 0x42, 0x0e, 0x70, 0xdf, 0x46, 0x8f, 0xd8, 0x63, 0xf2, 0xe8,
	0x2f, 0x51, 0x46, 0x8a, 0x77, 0x06, 0xe8, 0x6d, 0xf6, 0x8c, 0x3c, 0xd9, 0x1c, 0x3f, 0x56, 0xa1,
	0x88, 0x2e, 0xf7, 0x03, 0xdc, 0x8d, 0xe9, 0x1d, 0xf6, 0x8a, 0xbc, 0x4c, 0xc5, 0x18, 0x50, 0x2b,
	0x85, 0x43, 0x9e, 0x0c, 0x00, 0x23, 0x95, 0x60, 0x6c, 0x86, 0x5a, 0x6c, 0x5e, 0x65, 0xe7, 0xa1,
	0xa4, 0xf7, 0xf3, 0x80, 0x74, 0x3e, 0xda, 0xb9, 0xf7, 0xff, 0x3a, 0x7b, 0x6c, 0xaf, 0x97, 0x51,
	0x53, 0xe2, 0xe8, 0x60, 0x1c, 0x6e, 0x5c, 0xa5, 0xbd, 0x9a, 0x56, 0xa5, 0x67, 0x97, 0xa5, 0x5f,
	0x16, 0xd5, 0xba, 0xe2, 0xed, 0x2a, 0x2d, 0x66, 0xf5, 0xbf, 0x36, 0xeb, 0xad, 0xfb, 0x7c, 0x6d,
	0xb5, 0x07, 0x9c, 0x7f, 0x6b, 0x9d, 0x0d, 0x5c, 0x18, 0xcf, 0x6b, 0xcf, 0xc1, 0x06, 0x65, 0x5d,
	0x6f, 0x3d, 0xb2, 0xfe, 0xbe, 0x15, 0x4c, 0x78, 0x5e, 0x4f, 0x76, 0x82, 0x49, 0xd6, 0x9d, 0x38,
	0xc1, 0x8f, 0x56, 0xc7, 0xb1, 0x41, 0xc0, 0xf3, 0x3a, 0x08, 0x76, 0x92, 0x20, 0xc8, 0xba, 0x41,
	0xe0, 0x44, 0x1f, 0x0e, 0xd7, 0xa7, 0x7b, 0xfd, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x59, 0x09, 0xd2,
	0xb2, 0xf6, 0x02, 0x00, 0x00,
}
