package ccmsg

import (
	"errors"
	fmt "fmt"
	"math"
	"net"

	proto "github.com/golang/protobuf/proto"
)

func (m *EscrowInfo) TotalBlocks() uint64 {
	var x uint64
	for _, seg := range m.TicketsPerBlock {
		x += seg.Length
	}
	return x
}

func (m *EscrowInfo) TotalTickets() uint64 {
	var x uint64
	for _, seg := range m.TicketsPerBlock {
		x += seg.Value
	}
	return x
}

func (m *EscrowInfo) TicketsInBlock(block uint64) uint64 {
	for _, seg := range m.TicketsPerBlock {
		if block < seg.Length {
			return seg.Value
		}
		block -= seg.Length
	}
	return 0
}

func (m *TicketBundle) BuildClientCacheRequest(subMsg proto.Message) (*ClientCacheRequest, error) {
	r := &ClientCacheRequest{
		BundleRemainder:        m.Remainder,
		TicketBundleSubdigests: m.GetSubdigests(),
		BundleSig:              m.BatchSig,
		BundleSignerCert:       m.BundleSignerCert,
	}

	switch subMsg := subMsg.(type) {
	case (*TicketRequest):
		r.Ticket = &ClientCacheRequest_TicketRequest{TicketRequest: subMsg}
	case (*TicketL1):
		r.Ticket = &ClientCacheRequest_TicketL1{TicketL1: subMsg}
	case (*TicketL2Info):
		r.Ticket = &ClientCacheRequest_TicketL2{TicketL2: subMsg}
	default:
		return nil, errors.New("unexpected submessage type")
	}

	return r, nil
}

func (a *NetworkAddress) ConnectionString() string {
	if len(a.Inetaddr) > 0 {
		return fmt.Sprintf("%v:%v", net.IP(a.Inetaddr), a.Port)
	}
	if len(a.Inet6Addr) > 0 {
		return fmt.Sprintf("%v:%v", net.IP(a.Inet6Addr), a.Port)
	}
	return ""
}

func (m *ObjectMetadata) BlockCount() uint64 {
	return uint64(math.Ceil(float64(m.ObjectSize) / float64(m.BlockSize)))
}
