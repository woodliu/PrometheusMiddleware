package exporter

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "net/http"
    "runtime"
    "github.com/woodliu/PrometheusMiddleware/pkg/common"
    "sync/atomic"
    "time"

    "go.opentelemetry.io/otel/api/global"
    "go.opentelemetry.io/otel/api/metric"
    "go.opentelemetry.io/otel/exporters/metric/prometheus"
    "go.opentelemetry.io/otel/label"
)

const (
    ScanPeriod = time.Second * 30 //TODO: 扫描周期确定
)

/*
{
	"metrics": [{
			"name": "metric_1",
			"labels": [{
					"key": "metric_1_key1",
					"value": "metric_1_value1"
				},
			"description": "for metric_1"
		},
		{
			"name": "metric_2",
			"labels": [{
					"key": "metric_2_key1",
					"value": "metric_2_value1"
				},
				{
					"key": "metric_2_key2",
					"value": "metric_2_value2"
				}
			],
			"description": "for metric_2"
		}
	]
}
*/
type Metric struct {
    Name string         `json:"name"`           // metric 名称
    Labels []Label      `json:"labels"`         // metric的label
    Description string  `json:"description"`    // 说明该metric的用途
    exportLabel []label.KeyValue
    count *int64
    new bool
}

type InputMetrics struct {
    Metrics []Metric     `json:"metrics"`
}

type Label struct {
    Key string          `json:"key"`
    Value string        `json:"value"`
}

type exportMetrics struct {
    ch chan InputMetrics
    metrcis map[string]*Metric
    //new bool /* 可以考虑在有新metric的时候置为1，这样在exporter的扫描时就不会全遍历，但这样需要加锁，导致两个goroutine的协同 */
}

var GlobalExporterMetrics *exportMetrics

func NewExporterMetric()*exportMetrics{
    return &exportMetrics{
        metrcis: make(map[string]*Metric),
        ch: make(chan InputMetrics),
    }
}

func NewMetric(name,description string,lables []Label) *Metric{
    var exportLabels []label.KeyValue
    for _,v := range lables{
        exportLabels = append(exportLabels, label.String(v.Key,v.Value))
    }

    return &Metric{
        Name: name,
        Labels: lables,
        Description: description,
        exportLabel: exportLabels,
        count: new(int64),
        new: true,
    }
}

func (e *exportMetrics)calculateMetric(m *InputMetrics){
    for _,v := range m.Metrics{
        if _,ok := e.metrcis[v.Name];!ok{
            e.metrcis[v.Name] = NewMetric(v.Name, v.Description, v.Labels)
        }

        atomic.AddInt64(e.metrcis[v.Name].count,1)
    }
}

func startCalculateInput(){
    /* 计算启动的goroutine数目,初始默认启用4个goroutine */
    leftGoroutineNum := 4
    /* 如果当前仅有1个CPU core，则只能启动一个goroutine */
    if runtime.NumCPU() == 1{
        leftGoroutineNum = 1
    /* 如果CPU core大于1，则需要给其他处理至少预留一个core，防止高流量导致其他处理阻塞 */
    }else if runtime.NumCPU() - 1 < leftGoroutineNum{
        leftGoroutineNum = runtime.NumCPU() - 1
    }

    for i := 0; i < leftGoroutineNum; i++ {
        go func() {
            for  {
                select {
                case inputMetric := <- GlobalExporterMetrics.ch:
                    GlobalExporterMetrics.calculateMetric(&inputMetric)
                }
            }
        }()
    }
}

//给总数即可，使用increase或rate计算即可
func Exporter(c *gin.Context){
    var inputMetric InputMetrics
    if err := c.ShouldBindJSON(&inputMetric);nil != err{
        c.JSON(http.StatusBadRequest, nil)
        logrus.WithFields(logrus.Fields{"Process":"Exporter"}).Errorln(common.InvalidParametersStructErr)
        return
    }

    /* 统计metric数 */
    GlobalExporterMetrics.ch <- inputMetric
    c.JSON(http.StatusOK, nil)
}

func initMeter() {
    exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
    if err != nil {
        logrus.WithFields(logrus.Fields{"Process":"Exporter"}).Panicf("failed to initialize prometheus exporter %v", err)
    }

    http.HandleFunc("/metrics", exporter.ServeHTTP)
    go func() {
        _ = http.ListenAndServe(":19091", nil)
    }()

    logrus.WithFields(logrus.Fields{"Process":"Exporter"}).Infoln("Prometheus server running on :19091")
}

func StartExporter(){
    GlobalExporterMetrics = NewExporterMetric()
    initMeter()
    startCalculateInput()

    go func() {
        meter := global.Meter("test.com")
        // 每隔一段时间检查一次是否有新增的metric
        ticker := time.NewTicker(ScanPeriod)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                for _,v := range GlobalExporterMetrics.metrcis{
                    if v.new{
                        cb := func(_ context.Context, result metric.Int64ObserverResult) {
                            value := atomic.LoadInt64(GlobalExporterMetrics.metrcis[v.Name].count)
                            labels := GlobalExporterMetrics.metrcis[v.Name].exportLabel
                            result.Observe(value, labels...)
                        }
                        _ = metric.Must(meter).NewInt64ValueObserver(v.Name,cb, metric.WithDescription(v.Description))
                        v.new = false
                    }
                }
            }
        }
    }()
}