module github.com/woodliu/PrometheusMiddleware

go 1.15

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.15.0
	github.com/sirupsen/logrus v1.7.0
	github.com/tidwall/gjson v1.6.3
	go.opentelemetry.io/otel v0.13.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.13.0
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
)
