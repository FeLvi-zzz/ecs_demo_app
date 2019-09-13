package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
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
		router := gin.Default()

		router.Use(func(ctx *gin.Context) {
			TraceSeg(ctx, ctx.Request.URL.Path) //ここで返ってくるcontextを次のハンドラに渡したい…
			ctx.Next()
		})

		router.GET("/",        notFoundHandler)
		router.GET("/hc",      healthHandler)
		router.GET("/info",    infoHandler)
		router.GET("/fibo",    fiboHandler)
		router.GET("/zipcode", zipcodeHandler)
		router.GET("/down",    downHandler)

		router.Run(":8080")

	}
}

func TraceSeg(c context.Context, service string) (*context.Context) {
	ctx, seg := xray.BeginSegment(c, service)
	fmt.Println(service)
	seg.Close(nil)

	return &ctx
}

func TraceSubSeg(c context.Context, service string) (*context.Context) {
	ctx, subSeg := xray.BeginSubsegment(c, service)
	fmt.Println(service)
	subSeg.Close(nil)

	return &ctx
}



func notFoundHandler(ctx *gin.Context) {
	fmt.Println("--- notFoundHandler")
	ctx.String(404, "404 Not Found!!")
}

func healthHandler(ctx *gin.Context) {
	fmt.Println("--- healthHandler")
	ctx.String(200, "OK")
}

func infoHandler(ctx *gin.Context) {
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
		ctx.String(500, "ERROR")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.String(500, "ERROR")
		return
	}
	var metadata interface{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		ctx.String(500, "ERROR")
		return
	}
	taskArn := metadata.(map[string]interface{})["Labels"].(map[string]interface{})["com.amazonaws.ecs.task-arn"].(string)
	task := strings.Split(taskArn, "/")[1]
	// レスポンス
	ctx.String(200, "instanceId: "+instanceId+"\ntask: "+task+"\ncontainerId: "+containerId)
}

func fiboHandler(ctx *gin.Context) {
	fmt.Println("--- fiboHandler")
	n, err := strconv.Atoi(ctx.Query("n"))
	if err != nil {
		ctx.String(500, "ERROR")
		return
	}
	ctx.String(200, strconv.Itoa(n)+"番目のフィボナッチ数は、"+strconv.Itoa(fibo(n)))
}

func zipcodeHandler(ctx *gin.Context) {
	fmt.Println("--- zipcodeHandler")
	newCtx := ctx.Request.Context()

	zipcode, err := strconv.Atoi(ctx.Query("zipcode"))
	if err != nil {
		ctx.String(500, "ERROR")
		return
	}

	myClient := xray.Client(http.DefaultClient)

	req, err := http.NewRequest(http.MethodGet, "http://zipcloud.ibsnet.co.jp/api/search?zipcode=" + strconv.Itoa(zipcode), nil)
	if err != nil {
		fmt.Errorf("[BUG] failed to build request: %s", err)
		return
	}

	// c := (*((*ctx).Request)).Context()
	// fmt.Println(c)

	resp, err := myClient.Do(req.WithContext(*newCtx))

	ctx.String(200, "http://zipcloud.ibsnet.co.jp/api/search?zipcode=" + strconv.Itoa(zipcode))
	if err != nil {
		ctx.String(500, "ERROR")
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
  if err == nil {
    ctx.String(200, string(b))
  }

}

func downHandler(ctx *gin.Context) {
	fmt.Println("--- downHandler")
	ctx.String(500, "DOWN!!!")
	log.Fatal("DOWN!!!")
}

func fibo(n int) int {
	if n < 2 {
		return 1
	}
	return fibo(n-2) + fibo(n-1)
}
