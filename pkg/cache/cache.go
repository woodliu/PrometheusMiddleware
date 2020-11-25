package cch

import (
	"github.com/patrickmn/go-cache"
	"time"
)

const(
	cacheExpiration = time.Second * 5 //TODO:确认此时间
	cleanupInterval = time.Second * 5
)

func AddCache(c *cache.Cache, k string, x interface{}) error {
	return c.Add(k, x, cacheExpiration)
}

func NewCache() *cache.Cache{
	/* 每cleanupInterval之后都会根据cache中item的过期时间进行一次清理，此处将所有item的过期时间都设置为CacheExpiration */
	return cache.New(cacheExpiration, cleanupInterval)
}