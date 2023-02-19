package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mongodb struct {
		URL        string
		Database   string
		Collection string
	}
	CacheRedis cache.CacheConf
}
