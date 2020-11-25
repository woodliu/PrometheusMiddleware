package config

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
	cch "github.com/woodliu/PrometheusMiddleware/pkg/cache"
	"sync"
	"time"
)

const(
	DuplicateMetricCfgErr = "Duplicate RawMetric in config.json"
	SampleLimit = 1000 //TODO:tiao zheng
)
var (
	ConfigFilePath = "D:\\config.json"  //TODO:使用下一行的内容
	//ConfigFilePath = "/root/config.json"
    LocalServerListener = "0.0.0.0:19090"
    QueryPath = "/prom/query"
    SearchTimeRange = time.Hour * 24 * 15
	EarliestDateDiff = time.Hour * 24 * 15
)

type (
	MetricConfigs struct{
		Basic
		MetricCfg []Config			`json:"metricCfg"`
	}

	RawMetricName string

	Basic struct {
		PrometheusUrl string		`json:"prometheusUrl"`
		Limit rate.Limit			`json:"limit"`
		Burst int					`json:"burst"`
	}

	Config struct {
		RawMetric RawMetricName		`json:"rawMetric"`
		RealMetric string			`json:"realMetric"`
		ExpectedResNum int			`json:"expectedResNum"`
	}
)

type (
	MetricMap struct {
		sync.Mutex
		Basic
		Map map[RawMetricName]*RealMetricInfo
		Cache *cache.Cache
		Limiter *rate.Limiter
	}

	RealMetricInfo struct {
		RealMetric string
		ExpectedResNum int
	}
)

var MetricMapConfig MetricMap

func LoadConfig(configFile string)error{
	var metricCfgs MetricConfigs

	file, err := os.Open(configFile)
	if nil != err{
		logrus.Errorf("Open config.json fail: %v\n", err)
		return err
	}

	cfg, err  := ioutil.ReadAll(file)
	if nil != err{
		logrus.Errorf("Read config.json fail: %v\n", err)
		return err
	}

	if err := json.Unmarshal(cfg, &metricCfgs);nil != err{
		logrus.Errorf("Unmarshal config.json fail: %v\n", err)
		return err
	}

	MetricMapConfig.Lock()

	/* 初始化基本信息 */
	MetricMapConfig.PrometheusUrl = metricCfgs.PrometheusUrl
	MetricMapConfig.Limiter =  rate.NewLimiter(metricCfgs.Limit, metricCfgs.Burst)
	MetricMapConfig.Cache = cch.NewCache()
	/* 初始化metric配置信息 */
	MetricMapConfig.Map = make(map[RawMetricName]*RealMetricInfo)
	for _,v := range metricCfgs.MetricCfg{
		MetricMapConfig.Map[v.RawMetric] = &RealMetricInfo{v.RealMetric,v.ExpectedResNum}
	}

	MetricMapConfig.Unlock()

	if len(MetricMapConfig.Map) != len(metricCfgs.MetricCfg){
		logrus.Errorf("Duplicate RawMetric in config.json\n")
		return errors.New(DuplicateMetricCfgErr)
	}

	return nil
}