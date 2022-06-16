/*
* @Author: wangqilong
* @Description:
* @File: ping
* @Date: 2021/9/23 10:40 上午
 */

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
)

const (
	ProtocolICMP = 1 // Internet Control Message
)

type Ping struct {
	SrcAddr    string
	DstAddr    string
	Timeout    int
	Count      int
	SuccSum    int
	LossRate   float64
	MinDelay   time.Duration
	MaxDelay   time.Duration
	AvgDelay   time.Duration
	TotalDelay time.Duration
}

func (p *Ping) listenForSpecific4(conn *icmp.PacketConn, neededPeer string, neededBody []byte, pid, needSeq int, sent []byte, mc *memCache) (string, []byte, error) {
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
	srcAddr := p.SrcAddr
	dstAddr := net.IPAddr{IP: net.ParseIP(p.DstAddr)}
	c, err := icmp.ListenPacket("ip4:icmp", srcAddr)
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

	for i := 0; i < p.Count; i++ {
		seq++
		t := time.Now()
		err = c.SetDeadline(time.Now().Add(time.Duration(p.Timeout) * time.Second))
		if err != nil {
			return err
		}

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
			return err
		}

		if _, err := c.WriteTo(wb, &dstAddr); err != nil {
			continue
		}

		_, _, err = p.listenForSpecific4(c, "", append(bs, 'x'), id, seq, wb, mc)
		if err != nil {
			continue
		}

		elapsed := time.Since(t)
		p.SuccSum++

		if p.MaxDelay == time.Duration(0) || elapsed > p.MaxDelay {
			p.MaxDelay = elapsed
		}

		if p.MinDelay == time.Duration(0) || elapsed > p.MinDelay {
			p.MinDelay = elapsed
		}

		p.TotalDelay += elapsed
		p.AvgDelay = time.Duration((int64)(p.TotalDelay/time.Microsecond)/(int64)(p.SuccSum)) * time.Microsecond
		time.Sleep(time.Second)
	}

	p.LossRate = float64((p.Count-p.SuccSum)/p.Count) * 100
	return nil
}

func NewPing(srcAddr string, dstAddr string, timeout int, count int) *Ping {
	return &Ping{
		SrcAddr: srcAddr,
		DstAddr: dstAddr,
		Timeout: timeout,
		Count:   count,
	}
}
