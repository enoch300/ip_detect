package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	RDB     *redis.Client
	Cluster *redis.ClusterClient
)

// Connect 初始化连接
func Connect() (err error) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "172.16.107.27:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 10, // 连接池大小
	})

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = RDB.Ping(c).Result()
	if err != nil {
		return err
	}

	return err
}

// ConnectCluster Connect 初始化连接
func ConnectCluster() (err error) {
	Cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{"192.168.221.180:6381"},
		PoolSize:     10,
		MinIdleConns: 2,
		PoolTimeout:  5 * time.Second,
	})

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = Cluster.Ping(c).Result()
	if err != nil {
		return err
	}
	return nil
}
