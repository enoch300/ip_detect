package detect

import (
	"github.com/go-redis/redis/v8"
	"github.com/pochard/commons/randstr"
	"github.com/vmihailenco/msgpack"
	"ip_detect/api"
	"ip_detect/utils/log"
	"ip_detect/utils/ping"
	"net"
	"os"
	"time"
)

type Msg struct {
	Dev       string `json:"dev"`
	Mid       string `json:"mid"`
	Biz       string `json:"biz"`
	BId       string `json:"bid"`
	BD        string `json:"bd"`
	Region    string `json:"region"`
	OuterIp   string `json:"outer_ip"`
	OuterPort string `json:"outer_port"`
	Ping      bool   `json:"ping"`
	Mtr       bool   `json:"mtr"`
	CheckPort bool   `json:"check_port"`
}

type Task struct {
	T          string
	Uid        string
	Target     *Msg
	PingReturn *ping.Ping
	PortAlive  int
}

func (t *Task) ping() {
	if !t.Target.Ping {
		return
	}

	p := ping.NewPing("0.0.0.0", t.Target.OuterIp, 3, 32)
	if err := p.SendICMP(); err != nil {
		log.GlobalLog.Errorf("ping %v -> %v %v", err, "0.0.0.0", t.Target.OuterIp)
		return
	}

	t.PingReturn = p
}

func (t *Task) port() {
	if t.Target.CheckPort {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(t.Target.OuterIp, t.Target.OuterPort), time.Duration(3)*time.Second)
		if err != nil {
			t.PortAlive = 0
		} else {
			if conn != nil {
				t.PortAlive = 1
				_ = conn.Close()
			} else {
				t.PortAlive = 0
			}
		}
	}
}

func (t *Task) Do(c chan struct{}) {
	t.ping()
	t.port()
	log.GlobalLog.Infof("监测时间: %s, 业务名: %v, 业务ID: %v, 业务BD:%v, 监控目标: %v:%s(%v), 归属: %s, 平均延时: %.2f, 最大延时: %.2f, 最小延时: %.2f, 丢包率: %.2f",
		t.T, t.Target.Biz, t.Target.BId, t.Target.BD, t.Target.OuterIp, t.Target.OuterPort, t.PortAlive, t.Target.Region, t.PingReturn.AvgDelay.Seconds()*1000, t.PingReturn.MaxDelay.Seconds()*1000, t.PingReturn.MinDelay.Seconds()*1000, t.PingReturn.LossRate)
	go PushToIpaas(t)
	<-c
}

func NewTask(m *redis.Message) (*Task, error) {
	var msg *Msg
	err := msgpack.Unmarshal([]byte(m.Payload), &msg)
	if err != nil {
		return nil, err
	}

	return &Task{
		T:          time.Now().Format("2006-01-02 15:04:05"),
		Uid:        randstr.RandomAlphanumeric(17),
		Target:     msg,
		PingReturn: &ping.Ping{},
	}, nil
}

func PushToIpaas(t *Task) {
	var values [][]interface{}
	hostname, _ := os.Hostname()
	value := []interface{}{t.T, t.Uid, t.Target.Mid, t.Target.Dev, t.Target.Biz, t.Target.BD, t.Target.BId, t.Target.Region, hostname, t.Target.OuterIp, t.Target.OuterPort, t.PortAlive, t.PingReturn.AvgDelay.Seconds() * 1000, t.PingReturn.MaxDelay.Seconds() * 1000, t.PingReturn.MinDelay.Seconds() * 1000, t.PingReturn.LossRate}
	values = append(values, value)
	api.PushToIpaas("ipaas", "ip_detect", []string{"t", "id", "mid", "device", "business", "bd", "bid", "region", "src", "dst", "dport", "dport_alive", "avg", "max", "min", "loss_rate"}, values)

	//上报MTR
	//if len(t.MtrReturn) == 0 {
	//	return
	//}
	//var mtrValues [][]interface{}
	//for _, h := range t.MtrReturn {
	//	value = []interface{}{t.T, t.Uid, strconv.Itoa(h.RouteNo), h.Addr, float64(h.Loss), strconv.Itoa(h.Snt), float64(h.Avg), float64(h.Best), float64(h.Wrst)}
	//	mtrValues = append(mtrValues, value)
	//}
	//
	//api.PushToIpaas("ipaas", "ip_detect_mtr", []string{"t", "id", "no", "host", "loss", "snt", "avg", "best", "wrst"}, mtrValues)
}
