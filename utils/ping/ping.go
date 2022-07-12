package ping

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"math"
	"math/rand"
	"net"
	"time"
)

const (
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
)

const (
	DefaultSrcAddr = "0.0.0.0"
	DefaultTimeout = 3 * time.Second
	DefaultCount   = 32
	DefaultTTL     = 128
)

type PingReturn struct {
	SuccSum    int
	LossRate   float64
	TotalDelay time.Duration
	MaxDelay   time.Duration
	MinDelay   time.Duration
	AvgDelay   time.Duration
}

type Ping struct {
	srcAddr  net.IPAddr
	dstAddr  net.IPAddr
	count    int
	ttl      int
	timeout  time.Duration
	interval time.Duration
}

type Option func(p *Ping)

func WithSrcAddr(src string) Option {
	return func(p *Ping) {
		p.srcAddr = net.IPAddr{IP: net.ParseIP(src)}
	}
}

func WithDstAddr(dst string) Option {
	return func(p *Ping) {
		p.dstAddr = net.IPAddr{IP: net.ParseIP(dst)}
	}
}

func WithCount(i int) Option {
	return func(p *Ping) {
		p.count = i
	}
}

func WithInterval(i int) Option {
	return func(p *Ping) {
		p.interval = time.Duration(i) * time.Millisecond
	}
}

func WithTimeout(i int) Option {
	return func(p *Ping) {
		p.timeout = time.Duration(i) * time.Second
	}
}

func WithTTL(i int) Option {
	return func(p *Ping) {
		p.ttl = i
	}
}

func (p *Ping) SrcAddr() string {
	return p.srcAddr.IP.String()
}

func (p *Ping) DstAddr() string {
	return p.dstAddr.IP.String()
}

func (p *Ping) Timeout() time.Duration {
	return p.timeout
}

func (p *Ping) Interval() time.Duration {
	return p.interval
}

func (p *Ping) Count() int {
	return p.count
}

func (p *Ping) listenForIpv4(c *icmp.PacketConn, neededPeer string, neededBody []byte, pid, needSeq int, sent []byte) (string, []byte, error) {
	for {
		buf := make([]byte, 1500)
		n, peer, err := c.ReadFrom(buf)
		if err != nil {
			if neterr, ok := err.(*net.OpError); ok && neterr.Timeout() {
				return "*", []byte{}, neterr
			}
		}

		if n == 0 {
			continue
		}

		if neededPeer != "" && peer.String() != neededPeer {
			continue
		}

		x, err := icmp.ParseMessage(ProtocolICMP, buf[:n])
		if err != nil {
			continue
		}

		if typ, ok := x.Type.(ipv4.ICMPType); ok && typ.String() == "time exceeded" {
			body := x.Body.(*icmp.TimeExceeded).Data
			index := bytes.Index(body, sent[:4])
			if index > 0 {
				x, _ = icmp.ParseMessage(ProtocolICMP, body[index:])
				switch x.Body.(type) {
				case *icmp.Echo:
					echoBody := x.Body.(*icmp.Echo)
					if echoBody.Seq == needSeq && echoBody.ID == pid {
						return peer.String(), []byte{}, nil
					}
					continue
				default:
					// ignore
				}
			}
		}

		if typ, ok := x.Type.(ipv4.ICMPType); ok && typ.String() == "echo reply" {
			b, _ := x.Body.Marshal(1)
			if string(b[4:]) != string(neededBody) {
				continue
			}
			echoBody := x.Body.(*icmp.Echo)
			if echoBody.Seq == needSeq && echoBody.ID == pid {
				return peer.String(), b[4:], nil
			}
			continue
		}
	}
}

func (p *Ping) PingIpv4() (pr *PingReturn, err error) {
	pingReturn := &PingReturn{}
	c, err := icmp.ListenPacket("ip4:icmp", p.srcAddr.IP.String())
	if err != nil {
		return pingReturn, err
	}
	defer c.Close()

	rand.Seed(time.Now().UnixNano())
	seq := rand.Intn(math.MaxUint16)
	id := rand.Intn(math.MaxUint16) & 0xffff

	for i := 0; i < p.Count(); i++ {
		seq++
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(seq))
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  seq,
				Data: append(bs, 'x'),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return pingReturn, err
		}

		_ = c.SetDeadline(time.Now().Add(p.Timeout()))
		t := time.Now()
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(p.DstAddr())}); err != nil {
			continue
		}

		_, _, err = p.listenForIpv4(c, "", append(bs, 'x'), id, seq, wb)
		if err != nil {
			continue
		}

		elapsed := time.Since(t)
		pingReturn.SuccSum++

		if pingReturn.MaxDelay == time.Duration(0) || elapsed > pingReturn.MaxDelay {
			pingReturn.MaxDelay = elapsed
		}

		if pingReturn.MinDelay == time.Duration(0) || elapsed < pingReturn.MinDelay {
			pingReturn.MinDelay = elapsed
		}

		pingReturn.TotalDelay += elapsed
		pingReturn.AvgDelay = time.Duration((int64)(pingReturn.TotalDelay/time.Microsecond)/(int64)(pingReturn.SuccSum)) * time.Microsecond
		time.Sleep(p.Interval())
	}

	pingReturn.LossRate = float64(p.Count()-pingReturn.SuccSum) / float64(p.Count()) * 100
	return pingReturn, err
}

func (p *Ping) listenForIpv6(c *icmp.PacketConn, neededPeer string, neededBody []byte, pid, needSeq int, sent []byte) (string, []byte, error) {
	for {
		buf := make([]byte, 1500)
		n, peer, err := c.ReadFrom(buf)
		if err != nil {
			if neterr, ok := err.(*net.OpError); ok && neterr.Timeout() {
				return "*", []byte{}, neterr
			}
		}

		if n == 0 {
			continue
		}

		if neededPeer != "" && peer.String() != neededPeer {
			continue
		}

		x, err := icmp.ParseMessage(ProtocolIPv6ICMP, buf[:n])
		if err != nil {
			continue
		}

		if x.Type.(ipv6.ICMPType) == ipv6.ICMPTypeTimeExceeded {
			body := x.Body.(*icmp.TimeExceeded).Data
			x, _ = icmp.ParseMessage(ProtocolIPv6ICMP, body[40:])
			switch x.Body.(type) {
			case *icmp.Echo:
				echoBody := x.Body.(*icmp.Echo)
				if echoBody.Seq == needSeq && echoBody.ID == pid {
					return peer.String(), []byte{}, nil
				}
				continue
			default:
				// ignore
			}

		}

		if typ, ok := x.Type.(ipv6.ICMPType); ok && typ == ipv6.ICMPTypeEchoReply {
			b, _ := x.Body.Marshal(1)
			if string(b[4:]) != string(neededBody) || x.Body.(*icmp.Echo).ID != pid {
				continue
			}

			return peer.String(), b[4:], nil
		}
	}
}

func (p *Ping) PingIpv6() (pr *PingReturn, err error) {
	pingReturn := &PingReturn{}
	c, err := icmp.ListenPacket("ip6:ipv6-icmp", p.SrcAddr())
	if err != nil {
		return pingReturn, err
	}
	defer c.Close()

	rand.Seed(time.Now().UnixNano())
	seq := rand.Intn(math.MaxUint16)
	id := rand.Intn(math.MaxUint16) & 0xffff

	for i := 0; i < p.Count(); i++ {
		seq++
		fmt.Println(i)
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(seq))
		wm := icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest,
			Code: 0,
			Body: &icmp.Echo{
				ID: id, Seq: seq,
				Data: append(bs, 'x'),
			},
		}

		wb, err := wm.Marshal(nil)
		if err != nil {
			return pingReturn, err
		}

		t := time.Now()
		_ = c.SetDeadline(time.Now().Add(p.Timeout()))
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(p.DstAddr())}); err != nil {
			continue
		}

		_, _, err = p.listenForIpv6(c, "", append(bs, 'x'), id, seq, wb)
		if err != nil {
			continue
		}

		elapsed := time.Since(t)
		pingReturn.SuccSum++
		if pingReturn.MaxDelay == time.Duration(0) || elapsed > pingReturn.MaxDelay {
			pingReturn.MaxDelay = elapsed
		}

		if pingReturn.MinDelay == time.Duration(0) || elapsed < pingReturn.MinDelay {
			pingReturn.MinDelay = elapsed
		}

		pingReturn.TotalDelay += elapsed
		pingReturn.AvgDelay = time.Duration((int64)(pingReturn.TotalDelay/time.Microsecond)/(int64)(pingReturn.SuccSum)) * time.Microsecond
		time.Sleep(p.Interval())
	}

	pingReturn.LossRate = float64(p.Count()-pingReturn.SuccSum) / float64(p.Count()) * 100
	return pingReturn, err
}

func NewPing(opts ...Option) *Ping {
	options := &Ping{
		srcAddr: net.IPAddr{IP: net.ParseIP(DefaultSrcAddr)},
		timeout: DefaultTimeout,
		count:   DefaultCount,
		ttl:     DefaultTTL,
	}

	for _, o := range opts {
		o(options)
	}

	return options
}
