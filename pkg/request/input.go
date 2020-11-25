package request

import (
	"github.com/woodliu/PrometheusMiddleware/config"
	"time"
)

/*
	{
	  "metric": "ISUP",
	  "sampleNum": 500,
	  "startTime": 1605193552,
	  "endTime": 1605193582
	}
*/
type Input struct{
	Metric config.RawMetricName		`json:"metric"`
	SampleNum int64					`json:"sampleNum"`
	StartTime int64					`json:"startTime"`
	EndTime int64					`json:"endTime"`
}

type StdResponse struct {
	Data        *Ress  `json:"data"`
	ErrMsg      string `json:"errmsg"`
}

var Precision map[string]time.Duration