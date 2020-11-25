package api

import (
	"github.com/patrickmn/go-cache"
	"github.com/woodliu/PrometheusMiddleware/config"
)

type Request interface {
	DoQuery(c *config.Config, cache *cache.Cache)(bool, error)
}