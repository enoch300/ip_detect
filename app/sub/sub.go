package sub

import (
	"context"
	"ip_detect/app/detect"
	"ip_detect/db/redis"
	"ip_detect/utils/log"
)

func Sub(channel string) {
	sub := redis.RDB.Subscribe(context.Background(), channel)
	_, err := sub.Receive(context.Background())
	if err != nil {
		log.GlobalLog.Errorf("SubMessage %v", err.Error())
		return
	}

	c := make(chan struct{}, 500)
	for msg := range sub.Channel() {
		task, err := detect.NewTask(msg)
		if err != nil {
			log.GlobalLog.Errorf("NewTask %v", err.Error())
			continue
		}
		c <- struct{}{}
		go task.Do(c)
	}
}
