package sub

import (
	"context"
	"encoding/json"
	"github.com/vmihailenco/msgpack"
	"ip_detect/api"
	"ip_detect/app/detect"
	"ip_detect/db/redis"
	"ip_detect/utils/log"
)

func SubMessage(channel string) {
	sub := redis.RDB.Subscribe(context.Background(), channel)
	_, err := sub.Receive(context.Background())
	if err != nil {
		log.GlobalLog.Errorf("SubMessage %v", err.Error())
		return
	}

	ch := sub.Channel()
	for msg := range ch {
		var target api.Target
		err := msgpack.Marshal([]byte(msg.Payload), &target)
		if err := json.Unmarshal([]byte(msg.Payload), &target); err != nil {
			log.GlobalLog.Errorf("json unmarshal %v", err.Error())
			continue
		}

		task := detect.NewTask(&target)
		go task.Detect()
	}
}
