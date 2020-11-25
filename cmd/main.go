package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"github.com/woodliu/PrometheusMiddleware/config"
	"github.com/woodliu/PrometheusMiddleware/pkg/common"
	"github.com/woodliu/PrometheusMiddleware/pkg/exporter"
	"github.com/woodliu/PrometheusMiddleware/pkg/request"
)

/*
	TODO:
    1: 完成代码中所有todo的内容
	2：增加日志打印，是request阶段还是exporter阶段

http首部加入 "Connection: keep-alive"
需要处理ExpectedResNum不一致的问题


	Pod 相关字段
	CPU：
	container_cpu_usage_seconds_total

	内存：
	container_memory_rss

	网络：
	container_network_receive_bytes_total  接收
	container_network_transmit_bytes_total 发送

	Isito 网关 2xx 和 4xx
	istio_requests_total 接收和发送 都是这个

	istio_request_bytes_bucket  接收的字节
	istio_response_bytes_bucket  发送出去字节
*/

func main(){
	if nil != common.InitEnv(){
		log.Panicln("Init Env err!")
	}
	/* 启动exporter */
	exporter.StartExporter()

	//下面禁用gin的request输出，用于提升性能
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	})

	r.POST(config.QueryPath, request.Process)
	r.POST("/reload", request.ReloadConfig)
	r.POST("/exporter", exporter.Exporter)
	err := r.Run(config.LocalServerListener)
	if err != nil {
		log.Panic(err)
	}
}
