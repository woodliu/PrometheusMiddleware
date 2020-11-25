package exporter

import (
    "fmt"
    "net/http"
    "strings"
    "testing"
)

var localExporterService = "http://127.0.0.1:19090/exporter"

/* 每次发送一个metric,GOMAXPROCS=8，测试结果如下：
    测试次数	总计次数	每次操作耗时(ns)
    1	     4812	236897
    2	     4296	239118
    3	     4296	237957
    4	     4285	236545
    5	     5100	239164
*/
func BenchmarkExporter1(b *testing.B) {
    bodyData := "{\n\t\"metrics\": [{\n\t\t\t\"name\": \"metric_1\",\n\t\t\t\"labels\": [{\n\t\t\t\t\t\"key\": \"metric_1_key1\",\n\t\t\t\t\t\"value\": \"metric_1_value1\"\n\t\t\t\t},\n\t\t\t\t{\n\t\t\t\t\t\"key\": \"metric_1_key2\",\n\t\t\t\t\t\"value\": \"metric_1_value2\"\n\t\t\t\t},\n\t\t\t\t{\n\t\t\t\t\t\"key\": \"metric_1_key3\",\n\t\t\t\t\t\"value\": \"metric_1_value3\"\n\t\t\t\t}\n\t\t\t],\n\t\t\t\"description\": \"for metric_1\"\n\t\t}\n\t]\n}\n"
    b.ResetTimer()
    for i:=1;i<b.N;i++{
        http.DefaultClient.Post(localExporterService,"application/json",strings.NewReader(bodyData))
    }
}

/* 每次发送100个metric,GOMAXPROCS=8，测试结果如下：
    测试次数	总计次数	每次操作耗时(ns)
    1	     100	23219845
    2	     100	24047242
    3	     100	23550732
    4	     100	24065586
    5	     100	23656697
*/
func BenchmarkExporter2(b *testing.B) {
    var bodyDatas []string
    for i:=0;i<100;i++{
        newData := fmt.Sprintf("{\n\t\"metrics\": [{\n\t\t\t\"name\": \"metric_%d\",\n\t\t\t\"labels\": [{\n\t\t\t\t\t\"key\": \"metric_1_key1\",\n\t\t\t\t\t\"value\": \"metric_1_value1\"\n\t\t\t\t},\n\t\t\t\t{\n\t\t\t\t\t\"key\": \"metric_1_key2\",\n\t\t\t\t\t\"value\": \"metric_1_value2\"\n\t\t\t\t},\n\t\t\t\t{\n\t\t\t\t\t\"key\": \"metric_1_key3\",\n\t\t\t\t\t\"value\": \"metric_1_value3\"\n\t\t\t\t}\n\t\t\t],\n\t\t\t\"description\": \"for metric_1\"\n\t\t}\n\t]\n}\n",i)
        bodyDatas = append(bodyDatas,newData)
    }

    b.ResetTimer()
    for i:=1;i<b.N;i++{

        for _,v := range bodyDatas{
            //b.StartTimer()
            http.DefaultClient.Post(localExporterService,"application/json",strings.NewReader(v))
          //  b.StopTimer()
        }
    }
}