package ping

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math"
	"math/rand"
	"net"
	"time"
	"xh_detect/utils/conv"
)

const (
	ProtocolICMP = 1 // Internet Control Message
)

type Ping struct {
	SrcAddr      string
	DstAddr      string
	Timeout      int
	SuccessTimes int
	FailTimes    int
	LossRate     float64
	MinDelay     float64
	MaxDelay     float64
	AvgDelay     float64
	TotalDelay   float64
}

func (p *Ping) listenForSpecific4(conn *icmp.PacketConn, deadline time.Time, neededPeer string, neededBody []byte, pid, needSeq int, sent []byte, mc *memCache) (string, []byte, error) {
	for {
		copy(mc.B, mc.Buf)
		n, peer, err := conn.ReadFrom(mc.B)
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

		x, err := icmp.ParseMessage(ProtocolICMP, mc.B[:n])
		if err != nil {
			continue
		}

		if typ, ok := x.Type.(ipv4.ICMPType); ok && typ.String() == "time exceeded" {
			body := x.Body.(*icmp.TimeExceeded).Data

			index := bytes.Index(body, sent[:4])
			if index > 0 {
				x, _ := icmp.ParseMessage(ProtocolICMP, body[index:])
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

type memCache struct {
	Buf []byte
	B   []byte
}

func (p *Ping) SendICMP() error {
	c, err := icmp.ListenPacket("ip4:icmp", p.SrcAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	rand.Seed(time.Now().UnixNano())
	seq := rand.Intn(math.MaxUint16)
	id := rand.Intn(math.MaxUint16) & 0xffff

	mc := &memCache{
		Buf: make([]byte, 500),
		B:   make([]byte, 500),
	}

	for i := 1; i <= 32; i++ {
		seq++
		start := time.Now()
		err = c.SetDeadline(time.Now().Add(time.Duration(p.Timeout) * time.Second))
		if err != nil {
			return err
		}

		bs := make([]byte, 64)
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
			return err
		}

		dstAddr := net.IPAddr{IP: net.ParseIP(p.DstAddr)}
		if _, err := c.WriteTo(wb, &dstAddr); err != nil {
			p.FailTimes++
			continue
		}

		_, _, err = p.listenForSpecific4(c, time.Now().Add(time.Duration(p.Timeout)), "", append(bs, 'x'), id, seq, wb, mc)
		if err != nil {
			p.FailTimes++
			continue
		}

		elapsed := float64(time.Since(start) / 1000000)

		if p.MinDelay > elapsed {
			p.MinDelay = elapsed
		}

		if p.MaxDelay < elapsed {
			p.MaxDelay = elapsed
		}

		p.TotalDelay += elapsed
		p.SuccessTimes++
		//fmt.Printf("%v. %v -> %v %vms\n", i, p.SrcAddr, p.DstAddr, elapsed)
		time.Sleep(time.Second)
	}

	p.LossRate = conv.FormatFloat64(float64(p.FailTimes) / float64(p.SuccessTimes+p.FailTimes) * 100)

	if p.LossRate == 100 {
		p.MinDelay = 0
	} else {
		p.AvgDelay = conv.FormatFloat64(p.TotalDelay / float64(p.SuccessTimes))
	}

	return nil
}

func NewPing(srcAddr string, dstAddr string, timeout int) *Ping {
	return &Ping{
		SrcAddr:      srcAddr,
		DstAddr:      dstAddr,
		Timeout:      timeout,
		SuccessTimes: 0,
		FailTimes:    0,
		LossRate:     0,
		MinDelay:     math.MaxFloat64,
		MaxDelay:     0,
		AvgDelay:     0,
		TotalDelay:   0,
	}
}
