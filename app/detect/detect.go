package detect

import (
	"fmt"
	"github.com/enoch300/nt/mtr"
	"github.com/enoch300/nt/ping"
	"github.com/pochard/commons/randstr"
	"ip_detect/api"
	"ip_detect/utils/log"
	"net"
	"sync"
	"time"
)

type Task struct {
	T          string
	Uid        string
	Target     *api.Target
	PingReturn ping.PingReturn
	MtrReturn  []mtr.Hop
	PortAlive  int
}

func (t *Task) ping(wg *sync.WaitGroup) {
	defer wg.Done()
	if !t.Target.DoPing {
		return
	}
	_, pingReturn, err := ping.Ping("0.0.0.0", t.Target.OuterIp, 32, 1000, 1000)
	if err != nil {
		log.GlobalLog.Errorf("ping %v", err.Error())
		return
	}
	t.PingReturn = pingReturn
}

func (t *Task) mtr(wg *sync.WaitGroup) {
	defer wg.Done()
	if !t.Target.DoMtr {
		return
	}
}

func (t *Task) checkPort(wg *sync.WaitGroup) {
	defer wg.Done()
	if !t.Target.DoCheckPort {
		return
	}

	address := net.JoinHostPort(t.Target.OuterIp, t.Target.OuterPort)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
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

func (t *Task) Detect() {
	//rand.Seed(time.Now().UnixNano())
	//n := rand.Intn(300)
	//time.Sleep(time.Second * time.Duration(n))
	fmt.Println("detect >>> ", t.Target.OuterIp)
	//wg := &sync.WaitGroup{}
	//wg.Add(3)
	//go t.ping(wg)
	//go t.mtr(wg)
	//go t.checkPort(wg)
	//wg.Wait()

	//log.GlobalLog.Infof("监测时间: %s, 业务名: %v, 业务ID: %v, 业务BD:%v, 触发策略: %v, 监控目标: %v(%s), 平均延时: %.2f, 最大延时: %.2f, 最小延时: %.2f, 丢包率: %.2f",
	//	t.T, t.Target.Biz, t.Target.BId, t.Target.BD, "丢包率>5%", t.Target.OuterIp, t.Target.Region, t.PingReturn.AvgTime.Seconds()*1000, t.PingReturn.WrstTime.Seconds()*1000, t.PingReturn.BestTime.Seconds()*1000, t.PingReturn.DropRate)
	//PushToIpaas(t)
}

func NewTask(target *api.Target) *Task {
	return &Task{
		T:          time.Now().Format("2006-01-02 15:04:05"),
		Uid:        randstr.RandomAlphanumeric(17),
		Target:     target,
		PingReturn: ping.PingReturn{},
		MtrReturn:  make([]mtr.Hop, 0),
		PortAlive:  0,
	}
}

func PushToIpaas(t *Task) {
	var values [][]interface{}
	value := []interface{}{t.T, t.Uid, t.Target.Mid, t.Target.Dev, t.Target.Biz, t.Target.BD, t.Target.BId, t.Target.Region, "ali", t.Target.OuterIp, t.Target.OuterPort, t.PortAlive, t.PingReturn.AvgTime.Seconds() * 1000, t.PingReturn.WrstTime.Seconds() * 1000, t.PingReturn.BestTime.Seconds() * 1000, t.PingReturn.DropRate}
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
