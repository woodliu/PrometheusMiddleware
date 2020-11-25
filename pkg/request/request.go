package request

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"net/http"
	"github.com/woodliu/PrometheusMiddleware/config"
	"github.com/woodliu/PrometheusMiddleware/pkg/cache"
	"github.com/woodliu/PrometheusMiddleware/pkg/common"
	"strings"
	"time"
)

const(
	RequestTimeout = 10 * time.Second
)

type PromQuery struct{
	api v1.API
	Tp int
	Key string
	RawQuery string
	Ts time.Time /* for Query */

	R *v1.Range /* for QueryRange */

	Matches []string /* for Series */

	Label string /* for LabelName */
	StartTime time.Time
	EndTime time.Time

	ExpectNum int
	Response *PromResponse
}

type PromResponse struct {
	Resp interface{}
	Err error
}

func (promQuery *PromQuery)setApi(promAddr string) error{
	if client, err := api.NewClient(api.Config{Address: promAddr});nil != err{
		log.Errorf("Creating client: %v\n", err)
		return err
	}else{
		promQuery.api = v1.NewAPI(client)
		return nil
	}
}

func (promQuery *PromQuery)DoQuery(cache *cache.Cache) (interface{}, error){
	/* 首先查找cache */
	if obj,isExist := cache.Get(promQuery.Key);isExist {
		return obj, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	var obj interface{}
	/* 目前仅支持QueryRange */
	switch promQuery.Tp {
	case common.Query:
		val, warnings, err := promQuery.api.Query(ctx, promQuery.RawQuery, promQuery.Ts)
		if err != nil {
			log.WithFields(log.Fields{"Type":common.Query}).Errorf("Querying Prometheus: %v\n", err)
			return nil, err
		}

		if len(warnings) > 0 {
			log.WithFields(log.Fields{"Type":common.Query}).Warning(warnings)
		}

		obj,err = promQuery.Decode(val)
		if nil != err{
			return obj,err
		}
	case common.QueryRange:
		val, warnings, err := promQuery.api.QueryRange(ctx, promQuery.RawQuery, v1.Range{promQuery.R.Start, promQuery.R.End,promQuery.R.Step})
		if err != nil {
			log.WithFields(log.Fields{"Type":common.QueryRange}).Errorf("Querying Prometheus: %v\n", err)
			return nil, err
		}

		if len(warnings) > 0 {
			log.WithFields(log.Fields{"Type":common.QueryRange}).Warning(warnings)
		}

		obj,err = promQuery.Decode(val)
		if nil != err{
			return obj,err
		}
	case common.Series:
		val, warnings, err := promQuery.api.Series(ctx, promQuery.Matches, promQuery.StartTime, promQuery.EndTime)
		if err != nil {
			log.WithFields(log.Fields{"Type":common.Series}).Errorf("Querying Prometheus: %v\n", err)
			return nil, err
		}
		if len(warnings) > 0 {
			log.WithFields(log.Fields{"Type":common.Series}).Warning(warnings)
		}

		var res Ress
		for _,v1 := range val{
			for key,v2 := range v1{
				res.Result = append(res.Result, Res{string(key),string(v2)})
			}
		}
		obj = res
	case common.LabelName:
		val, warnings, err := promQuery.api.LabelValues(ctx, promQuery.Label, promQuery.StartTime, promQuery.EndTime)
		if err != nil {
			log.WithFields(log.Fields{"Type":common.LabelName}).Errorf("Querying Prometheus: %v\n", err)
			return nil, err
		}
		if len(warnings) > 0 {
			log.WithFields(log.Fields{"Type":common.LabelName}).Warning(warnings)
		}

		var res Ress
		for _,v1 := range val{
			for _,v2 := range v1{
				res.Result = append(res.Result, Res{"",string(v2)})
			}
		}
		obj = res
	default:
		log.Errorf("%v\n", common.InvalidQueryTypeErr)
		return nil, errors.New(common.InvalidQueryTypeErr)
	}

	err := cch.AddCache(cache, promQuery.Key, obj)
	if nil != err{
		log.Errorf("Adding cache: %v\n", err)
	}
	return obj,nil
}


func timeToSecond(t string) string {
	return strings.Split(t,".")[0]
}

func (promQuery *PromQuery)Decode(res model.Value) (*Ress,error){
	data := Ress{}
	tp := res.Type()
	switch tp {
	case model.ValVector:
		resVector := res.(model.Vector)

		if len(resVector) != promQuery.ExpectNum{
			return nil,errors.New("Unexpect result number")
		}

		for _,v := range resVector{
			data.Result = append(data.Result, Res{v.Value.String(),timeToSecond(v.Timestamp.String())})
		}
	case model.ValMatrix:
		resMatrix := res.(model.Matrix)

		if len(resMatrix) != promQuery.ExpectNum{
			return nil,errors.New("Unexpect result number")
		}

		for _,v1 := range resMatrix{
			for _,v2 := range v1.Values{
				data.Result = append(data.Result, Res{v2.Value.String(), timeToSecond(v2.Timestamp.String())})
			}
		}
	case model.ValScalar:
		resScalar := res.(*model.Scalar)
		data.Result = append(data.Result, Res{resScalar.Value.String(), timeToSecond(resScalar.Timestamp.String())})
	case model.ValString:
		resString := res.(*model.String)
		data.Result = append(data.Result, Res{resString.Value, timeToSecond(resString.Timestamp.String())})
	default:
	}

	return &data,nil
}

func isLabel(rawQuery string) bool{
	return strings.HasPrefix(rawQuery,"{")
}

func genKey(in Input) string{
	code,_ := json.Marshal(in)
	return b64.URLEncoding.EncodeToString(code)
}

func ReloadConfig(c *gin.Context){
	config.LoadConfig(config.ConfigFilePath)
	c.String(http.StatusOK, fmt.Sprintf("%s", "reload successfully!"))
}

func validateQuery(in *Input)error{
	if in.StartTime >= in.EndTime{
		return errors.New(common.InvalidTimestampErr)
	}

	if float64(time.Now().Unix() - in.StartTime) > config.EarliestDateDiff.Seconds() {
		return errors.New(common.TooEarlyTimestampErr)
	}

	if float64(in.EndTime - in.StartTime) > config.SearchTimeRange.Seconds() {
		return errors.New(common.InvalidTimeRangeErr)
	}

	if in.SampleNum > config.SampleLimit || in.SampleNum < 1 {
		return errors.New(common.InvalidSampleNumErr)
	}

	if _,ok := config.MetricMapConfig.Map[in.Metric];!ok{
		return errors.New(common.InvalidMetricErr)
	}

	return nil
}

func Process(c *gin.Context){
	config.MetricMapConfig.Lock()
	defer config.MetricMapConfig.Unlock()
	if config.MetricMapConfig.Limiter.Allow() == false{
		c.JSON(http.StatusTooManyRequests, StdResponse{
			ErrMsg: common.TooManyRequestsErr,
		})

		return
	}

	var input Input

	if err := c.ShouldBindJSON(&input);nil != err{
		c.JSON(http.StatusBadRequest, StdResponse{
			ErrMsg: common.InvalidParametersStructErr,
		})

		log.Errorf("Unmarshal request parameters err: %v\n", err)
		return
	}

	if err := validateQuery(&input);nil != err{
		c.JSON(http.StatusBadRequest, StdResponse{
			ErrMsg: err.Error(),
		})
		return
	}

	rawQuery := config.MetricMapConfig.Map[input.Metric].RealMetric
	if isLabel(rawQuery){
		c.JSON(http.StatusBadRequest, StdResponse{
			ErrMsg: common.InvalidMetricErr,
		})
		return
	}

	/* 目前仅支持queryRange模式 */
	tp := common.QueryRange
	key := genKey(input)
	expectNum := config.MetricMapConfig.Map[input.Metric].ExpectedResNum

	var promQuery *PromQuery

	switch tp {
	case common.Query:
		promQuery = &PromQuery{
			Tp: tp,
			Key: key,
			RawQuery: rawQuery,
			Ts: time.Now(),
		}
	case common.QueryRange:
		promQuery = &PromQuery{
			Tp: tp,
			Key: key,
			RawQuery: rawQuery,
			R: &v1.Range{
				/* 使用以秒计数的时间 */
				time.Unix(input.StartTime,0),
				time.Unix(input.EndTime,0),
				/* 计算方式参考grafana: https://github.com/grafana/grafana/blob/master/pkg/tsdb/interval.go#L56 */
				time.Duration((input.EndTime - input.StartTime) / input.SampleNum)*1000000000,
			},
			ExpectNum: expectNum,
		}
	default:
		c.JSON(http.StatusBadRequest, StdResponse{
			ErrMsg: common.InvalidQueryTypeErr,
		})
		return
	}

	if err := promQuery.setApi(config.MetricMapConfig.PrometheusUrl); nil != err{
		c.JSON(http.StatusInternalServerError, StdResponse{
			ErrMsg: common.ForbiddenMetricErr,
		})

		log.WithFields(log.Fields{"Type":common.Query}).Errorf("Creating Prometheus: %v\n", err)
		return
	}

	res,err := promQuery.DoQuery(config.MetricMapConfig.Cache)
	if nil != err{
		c.JSON(http.StatusInternalServerError, StdResponse{
			ErrMsg: common.QueryPrometheusErr,
		})

		return
	}

	data := res.(*Ress)
	if int64(len(data.Result)) > input.SampleNum{
		data.Result = data.Result[:input.SampleNum]
	}

	c.JSON(http.StatusOK, StdResponse{
		Data: data,
	})
	return
}