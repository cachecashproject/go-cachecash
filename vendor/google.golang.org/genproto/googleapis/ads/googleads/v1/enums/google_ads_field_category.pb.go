// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/enums/google_ads_field_category.proto

package enums // import "google.golang.org/genproto/googleapis/ads/googleads/v1/enums"

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

// The category of the artifact.
type GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory int32

const (
	// Unspecified
	GoogleAdsFieldCategoryEnum_UNSPECIFIED GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 0
	// Unknown
	GoogleAdsFieldCategoryEnum_UNKNOWN GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 1
	// The described artifact is a resource.
	GoogleAdsFieldCategoryEnum_RESOURCE GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 2
	// The described artifact is a field and is an attribute of a resource.
	// Including a resource attribute field in a query may segment the query if
	// the resource to which it is attributed segments the resource found in
	// the FROM clause.
	GoogleAdsFieldCategoryEnum_ATTRIBUTE GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 3
	// The described artifact is a field and always segments search queries.
	GoogleAdsFieldCategoryEnum_SEGMENT GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 5
	// The described artifact is a field and is a metric. It never segments
	// search queries.
	GoogleAdsFieldCategoryEnum_METRIC GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory = 6
)

var GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	2: "RESOURCE",
	3: "ATTRIBUTE",
	5: "SEGMENT",
	6: "METRIC",
}
var GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory_value = map[string]int32{
	"UNSPECIFIED": 0,
	"UNKNOWN":     1,
	"RESOURCE":    2,
	"ATTRIBUTE":   3,
	"SEGMENT":     5,
	"METRIC":      6,
}

func (x GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory) String() string {
	return proto.EnumName(GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory_name, int32(x))
}
func (GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_google_ads_field_category_964e3bf29a6fd9b8, []int{0, 0}
}

// Container for enum that determines if the described artifact is a resource
// or a field, and if it is a field, when it segments search queries.
type GoogleAdsFieldCategoryEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GoogleAdsFieldCategoryEnum) Reset()         { *m = GoogleAdsFieldCategoryEnum{} }
func (m *GoogleAdsFieldCategoryEnum) String() string { return proto.CompactTextString(m) }
func (*GoogleAdsFieldCategoryEnum) ProtoMessage()    {}
func (*GoogleAdsFieldCategoryEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_google_ads_field_category_964e3bf29a6fd9b8, []int{0}
}
func (m *GoogleAdsFieldCategoryEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GoogleAdsFieldCategoryEnum.Unmarshal(m, b)
}
func (m *GoogleAdsFieldCategoryEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GoogleAdsFieldCategoryEnum.Marshal(b, m, deterministic)
}
func (dst *GoogleAdsFieldCategoryEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GoogleAdsFieldCategoryEnum.Merge(dst, src)
}
func (m *GoogleAdsFieldCategoryEnum) XXX_Size() int {
	return xxx_messageInfo_GoogleAdsFieldCategoryEnum.Size(m)
}
func (m *GoogleAdsFieldCategoryEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_GoogleAdsFieldCategoryEnum.DiscardUnknown(m)
}

var xxx_messageInfo_GoogleAdsFieldCategoryEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*GoogleAdsFieldCategoryEnum)(nil), "google.ads.googleads.v1.enums.GoogleAdsFieldCategoryEnum")
	proto.RegisterEnum("google.ads.googleads.v1.enums.GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory", GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory_name, GoogleAdsFieldCategoryEnum_GoogleAdsFieldCategory_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/enums/google_ads_field_category.proto", fileDescriptor_google_ads_field_category_964e3bf29a6fd9b8)
}

var fileDescriptor_google_ads_field_category_964e3bf29a6fd9b8 = []byte{
	// 336 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x50, 0xcf, 0x4a, 0xc3, 0x30,
	0x18, 0xb7, 0x1d, 0x4e, 0xcd, 0x14, 0x4b, 0x0f, 0x1e, 0xa6, 0x3b, 0x6c, 0x0f, 0x90, 0x52, 0xbc,
	0x45, 0x3c, 0xb4, 0x35, 0x1b, 0x45, 0xd6, 0x8d, 0xae, 0x9d, 0x20, 0x85, 0x11, 0x97, 0x1a, 0x0a,
	0x5d, 0x32, 0x96, 0x6e, 0xe0, 0x2b, 0xf8, 0x18, 0x1e, 0x7d, 0x14, 0x1f, 0xc5, 0x93, 0x8f, 0x20,
	0x4d, 0xd7, 0x9e, 0xa6, 0x97, 0x10, 0xbe, 0xdf, 0x9f, 0xef, 0xfb, 0xfd, 0xc0, 0x3d, 0x13, 0x82,
	0xe5, 0xa9, 0x45, 0xa8, 0xb4, 0xaa, 0x6f, 0xf9, 0xdb, 0xd9, 0x56, 0xca, 0xb7, 0xab, 0x7a, 0xb4,
	0x20, 0x54, 0x2e, 0x5e, 0xb3, 0x34, 0xa7, 0x8b, 0x25, 0x29, 0x52, 0x26, 0x36, 0x6f, 0x70, 0xbd,
	0x11, 0x85, 0x30, 0x7b, 0x15, 0x01, 0x12, 0x2a, 0x61, 0x23, 0x87, 0x3b, 0x1b, 0x2a, 0x79, 0xf7,
	0xa6, 0x76, 0x5f, 0x67, 0x16, 0xe1, 0x5c, 0x14, 0xa4, 0xc8, 0x04, 0x97, 0x95, 0x78, 0xf0, 0xae,
	0x81, 0xee, 0x48, 0x11, 0x1c, 0x2a, 0x87, 0xa5, 0xbd, 0xb7, 0x77, 0xc7, 0x7c, 0xbb, 0x1a, 0xe4,
	0xe0, 0xea, 0x30, 0x6a, 0x5e, 0x82, 0x4e, 0x1c, 0xcc, 0xa6, 0xd8, 0xf3, 0x87, 0x3e, 0x7e, 0x30,
	0x8e, 0xcc, 0x0e, 0x38, 0x89, 0x83, 0xc7, 0x60, 0xf2, 0x14, 0x18, 0x9a, 0x79, 0x0e, 0x4e, 0x43,
	0x3c, 0x9b, 0xc4, 0xa1, 0x87, 0x0d, 0xdd, 0xbc, 0x00, 0x67, 0x4e, 0x14, 0x85, 0xbe, 0x1b, 0x47,
	0xd8, 0x68, 0x95, 0xcc, 0x19, 0x1e, 0x8d, 0x71, 0x10, 0x19, 0xc7, 0x26, 0x00, 0xed, 0x31, 0x8e,
	0x42, 0xdf, 0x33, 0xda, 0xee, 0x8f, 0x06, 0xfa, 0x4b, 0xb1, 0x82, 0xff, 0x06, 0x72, 0xaf, 0x0f,
	0x5f, 0x34, 0x2d, 0xf3, 0x4c, 0xb5, 0x67, 0x77, 0xaf, 0x66, 0x22, 0x27, 0x9c, 0x41, 0xb1, 0x61,
	0x16, 0x4b, 0xb9, 0x4a, 0x5b, 0xb7, 0xbb, 0xce, 0xe4, 0x1f, 0x65, 0xdf, 0xa9, 0xf7, 0x43, 0x6f,
	0x8d, 0x1c, 0xe7, 0x53, 0xef, 0x55, 0x9b, 0xa0, 0x43, 0x25, 0x6c, 0x96, 0xc2, 0xb9, 0x0d, 0xcb,
	0x6e, 0xe4, 0x57, 0x8d, 0x27, 0x0e, 0x95, 0x49, 0x83, 0x27, 0x73, 0x3b, 0x51, 0xf8, 0xb7, 0xde,
	0xaf, 0x86, 0x08, 0x39, 0x54, 0x22, 0xd4, 0x30, 0x10, 0x9a, 0xdb, 0x08, 0x29, 0xce, 0x4b, 0x5b,
	0x1d, 0x76, 0xfb, 0x1b, 0x00, 0x00, 0xff, 0xff, 0xba, 0xef, 0x9f, 0x9b, 0x04, 0x02, 0x00, 0x00,
}
