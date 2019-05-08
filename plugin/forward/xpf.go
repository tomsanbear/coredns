package forward

import (
	"net"
	"strconv"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// HandleXPF adds in the XPF record if not present already
// TODO: at the moment we are overwriting any existing proxy records, as an additional implementation for IP range white listing
func HandleXPF(state *request.Request) error {
	// Discard an already present record
	for index, rr := range state.Req.Extra {
		_, ok := rr.(*dns.XPF)
		if ok {
			state.Req.Extra = append(state.Req.Extra[:index], state.Req.Extra[index+1:]...)
		}
	}

	// Insert the XPF record
	xpfRR := &dns.XPF{}

	if ipVersion := net.ParseIP(state.LocalIP()).To4(); ipVersion != nil {
		xpfRR.IpVersion = 4
		xpfRR.SrcAddress = net.ParseIP(state.IP()).To4()
		xpfRR.DestAddress = net.ParseIP(state.LocalIP()).To4()
	} else if ipVersion := net.ParseIP(state.LocalIP()).To16(); ipVersion != nil {
		xpfRR.IpVersion = 6
		xpfRR.SrcAddress = net.ParseIP(state.IP()).To16()
		xpfRR.DestAddress = net.ParseIP(state.LocalIP()).To16()
	}

	srcPort64, err := strconv.ParseUint(state.Port(), 16, 16)
	if err != nil {
		// TODO: Handle it
	}
	xpfRR.SrcPort = uint16(srcPort64)

	destPort64, err := strconv.ParseUint(state.LocalPort(), 16, 16)
	if err != nil {
		// TODO: Handle it
	}
	xpfRR.DestPort = uint16(destPort64)

	xpfRR.Protocol = protoIANA(state.Proto())

	// Append to the Additional Section
	state.Req.Extra = append(state.Req.Extra, xpfRR)

	xpfRR.Hdr = dns.RR_Header{
		Name:   ".",
		Rrtype: dns.TypeXPF,
		Class:  1,
		Ttl:    0,
	}

	clog.Infof("appended: %v", xpfRR)

	return nil
}

func protoIANA(proto string) uint8 {
	switch proto {
	case "udp":
		return 17
	case "tcp":
		return 6
	}
	return 17 // TODO: should error here?
}
