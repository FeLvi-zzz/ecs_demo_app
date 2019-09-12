package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-xray-sdk-go/xray"
	_ "github.com/aws/aws-xray-sdk-go/plugins/ecs"
)

func init() {
  xray.Configure(xray.Config{
    DaemonAddr:       "127.0.0.1:2000",
		ServiceVersion:   "1.2.3",
	})
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "hc" {
		fmt.Println("ok")
	} else {
		fmt.Println("=== Application API Starting!!")
		http.Handle("/",        xray.Handler(xray.NewFixedSegmentNamer("index"),   http.HandlerFunc(notFoundHandler)))
		http.Handle("/hc",      xray.Handler(xray.NewFixedSegmentNamer("health"),  http.HandlerFunc(healthHandler)))
		http.Handle("/info",    xray.Handler(xray.NewFixedSegmentNamer("info"),    http.HandlerFunc(infoHandler)))
		http.Handle("/fibo",    xray.Handler(xray.NewFixedSegmentNamer("fibo"),    http.HandlerFunc(fiboHandler)))
		http.Handle("/zipcode", xray.Handler(xray.NewFixedSegmentNamer("zipcode"), http.HandlerFunc(zipcodeHandler)))
		http.Handle("/down",    xray.Handler(xray.NewFixedSegmentNamer("down"),    http.HandlerFunc(downHandler)))
		http.ListenAndServe(":8080", nil)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- notFoundHandler")
	w.WriteHeader(http.NotFound)
	fmt.Fprint(w, "404 Not Found")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- healthHandler")
	fmt.Fprint(w, "OK")
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- infoHandler")
	// インスタンスIDの取得
	sess := session.Must(session.NewSession())
	svc := ec2metadata.New(sess)
	doc, _ := svc.GetInstanceIdentityDocument()
	instanceId := doc.InstanceID
	// コンテナIDの取得
	containerId, _ := os.Hostname()
	// タスクの取得
	resp, err := http.Get(os.Getenv("ECS_CONTAINER_METADATA_URI"))
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}
	var metadata interface{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}
	taskArn := metadata.(map[string]interface{})["Labels"].(map[string]interface{})["com.amazonaws.ecs.task-arn"].(string)
	task := strings.Split(taskArn, "/")[1]
	// レスポンス
	fmt.Fprint(w, "instanceId: "+instanceId+"\ntask: "+task+"\ncontainerId: "+containerId)
}

func fiboHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- fiboHandler")
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}
	fmt.Fprint(w, strconv.Itoa(n)+"番目のフィボナッチ数は、"+strconv.Itoa(fibo(n)))
}

func zipcodeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- zipcodeHandler")

	ctx := r.Context()

	zipcode, err := strconv.Atoi(r.URL.Query().Get("zipcode"))
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}

	myClient := xray.Client(http.DefaultClient)

	req, err := http.NewRequest(http.MethodGet, "http://zipcloud.ibsnet.co.jp/api/search?zipcode=" + strconv.Itoa(zipcode), nil)
	if err != nil {
		fmt.Errorf("[BUG] failed to build request: %s", err)
		return
	}
	resp, err := myClient.Do(req.WithContext(ctx))

	fmt.Fprint(w, "http://zipcloud.ibsnet.co.jp/api/search?zipcode=" + strconv.Itoa(zipcode))
	if err != nil {
		fmt.Fprint(w, "ERROR")
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
  if err == nil {
    fmt.Fprint(w, string(b))
  }

}

func downHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- downHandler")
	log.Fatal("DOWN!!!")
}

func fibo(n int) int {
	if n < 2 {
		return 1
	}
	return fibo(n-2) + fibo(n-1)
}
