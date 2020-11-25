package request

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"github.com/woodliu/PrometheusMiddleware/pkg/common"
	"strings"
	"testing"
	"time"
)

type API struct {
	Client  *http.Client
	baseURL string
}

var localPrometheusService = "http://127.0.0.1:19090/prom/query"
var reloadUrl = "http://127.0.0.1:19090/reload"

var configFile = "D:\\config.json"
var reloadFile = "D:\\reload.json"
/*
	1：正常的访问测试，预期返回ok
	2：输入不存在的metric，预期返回错误
	3：输入格式不正确的metric，预期返回错误
	4：输入的采样率过大，预期返回错误
	5：输入的采样率为负值，预期返回错误
	6：输入的采样率为0，预期返回错误
	7：输入的采样率格式错误，预期返回错误
	8：输入的开始时间等于终止时间，预期返回错误
	9：输入的开始时间大于终止时间，预期返回错误
	10：输入的时间格式错误，含字母，预期返回错误
	11：输入的时间格式错误，待小数点，预期返回错误
	12：输入的开始时间过早，预期返回错误
	13：输入的开始时间和结束时间的查询范围过大，预期返回错误
	14：测试cache，大量相同的请求，预期成功
	15：返回的结果与预期的结果数量不一致
	16：reload测试，预期成功

*/

//注意：必须首先启动PrometheusService服务,且服务的config.json路径配置为test/config.json

func TestNormal(t *testing.T) {
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(gjson.Get(string(dataByte), "errmsg").String())
	if resp.StatusCode != http.StatusOK{
		t.Fail()
	}
}

func TestNoExistMetricFromAPI(t *testing.T) {
	bodyData := "{\n  \"metric\": \"NO_EXIST_API\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidMetricErr{
		t.Fail()
	}
}

func TestNoExistMetricInternal(t *testing.T) {
	bodyData := "{\n  \"metric\": \"NO_EXIST\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(gjson.Get(string(dataByte), "errmsg").String())
	if gjson.Get(string(dataByte), "errmsg").String() != common.QueryPrometheusErr || resp.StatusCode != http.StatusInternalServerError{
		t.Fail()
	}
}

func TestErrFormatMetric(t *testing.T) {
	bodyData := "{\n  \"metric\": \"ERR_METRIC_FORMAT\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.QueryPrometheusErr || resp.StatusCode != http.StatusInternalServerError{
		t.Fail()
	}
}

func TestSampleTooBig(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100000,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidSampleNumErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestSampleNegative(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": -1,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidSampleNumErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestSampleZero(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 0,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidSampleNumErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestSampleErrFormat(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 1-asd,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidParametersStructErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestStartTimeEquelEndTime(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 500,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605234258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidTimestampErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestStartTimeLarger(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 500,\n  \"startTime\": 1605234298,\n  \"endTime\": 1605234258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidTimestampErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestStartTimeErrFormat1(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": abocd,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidParametersStructErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestStartTimeErrFormat2(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258.666,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidParametersStructErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func TestStartTimeTooEarly(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": 1,\n  \"endTime\": 20\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.TooEarlyTimestampErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

/* 将config.go中的SearchTimeRange调整为10s，重启服务后运行本用例即可 */
func TestStartTimeRangeTooBig(t *testing.T){
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605235258,\n  \"endTime\": 1605235358\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidTimeRangeErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}
}

func BenchmarkProcess(b *testing.B) {
	bodyData := "{\n  \"metric\": \"OK\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	b.ResetTimer()
	b.N = 5000
	for i:=1;i<b.N;i++{
		http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	}
}

func TestUnexpectResNum(t *testing.T) {
	bodyData := "{\n  \"metric\": \"UNEXPECT_RESULT_NUM\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if gjson.Get(string(dataByte), "errmsg").String() != common.QueryPrometheusErr || resp.StatusCode != http.StatusInternalServerError{
		t.Fail()
	}
}

func TestReloadConfig(t *testing.T) {
	originalContent, _ := ioutil.ReadFile(configFile)
	newContent, _ := ioutil.ReadFile(reloadFile)

	ioutil.WriteFile(configFile,newContent,0666)
	defer func() {
		ioutil.WriteFile(configFile,originalContent,0666)
		http.DefaultClient.Post(reloadUrl,"application/json",nil)
	}()

	bodyData := "{\n  \"metric\": \"NEW_METRIC\",\n  \"sampleNum\": 100,\n  \"startTime\": 1605234258,\n  \"endTime\": 1605235258\n}"
	resp,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	dataByte, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if gjson.Get(string(dataByte), "errmsg").String() != common.InvalidMetricErr || resp.StatusCode != http.StatusBadRequest{
		t.Fail()
	}

	/* 执行reload */
	http.DefaultClient.Post(reloadUrl,"application/json",nil)
	/* 等待reload获取锁完成重载 */
	time.Sleep(time.Second)
	/* 再次查询 */
	resp2,_ := http.DefaultClient.Post(localPrometheusService,"application/json",strings.NewReader(bodyData))
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK{
		t.Fail()
	}
}