package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	flag.Parse()
	if flag.Arg(0) == "hc" {
		fmt.Println("ok")
	} else {
		fmt.Println("=== Application API Starting!!")
		http.HandleFunc("/hc", healthHandler)
		http.HandleFunc("/info", infoHandler)
		http.HandleFunc("/fibo", fiboHandler)
		http.HandleFunc("/down", downHandler)
		http.ListenAndServe(":8080", nil)
	}
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
