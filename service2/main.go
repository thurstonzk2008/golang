package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thurstonzk2008/httpserver/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Response 是一个标准的 http 返回的最小单元
type Response struct {
	Code   int         // http status code
	Header http.Header // response header
	Body   string      // response body
}

// httpHandle 处理请求的 handle func
type httpHandle func(request *http.Request) Response

// GetClientIP 解析客户端真实的IP地址
// 解析顺序遵循标准的 http 协议
// X-Real-IP --> X-Forwarded-For(first) --> RemoteAddr
// 增加对IP地址的校验，防止脏数据的产生
func GetClientIP(r *http.Request) string {
	var clientIp string

	switch {
	case r.Header.Get("X-Real-IP") != "":
		clientIp = r.Header.Get("X-Real-IP")
	case r.Header.Get("X-Forwarded-For") != "":
		clientIps := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		clientIp = strings.TrimSpace(clientIps[0])
	}

	// 解析IP
	if net.ParseIP(clientIp) != nil {
		return clientIp
	} else {
		// default
		clientIp = strings.Split(r.RemoteAddr, ":")[0]
	}
	return clientIp
}

// httpHandleFunc 封装日志中间件
func httpHandleFunc(w http.ResponseWriter, r *http.Request,
	handle func(request *http.Request) Response) {
	beginTime := time.Now()
	// 随机耗时
	time.Sleep(time.Second * time.Duration(rand.Int31n(3)))

	// 解析客户端IP
	clientIp := GetClientIP(r)

	// before
	log.Printf("request_in||client_ip=%s||uri=%s\n", clientIp, r.RequestURI)

	response := handle(r)
	if response.Code == 0 {
		response.Code = http.StatusOK
	}
	duration := time.Since(beginTime).Seconds()

	// set header
	for s := range response.Header {
		for _, value := range response.Header[s] {
			w.Header().Add(s, value)
		}
	}
	// Code
	w.WriteHeader(response.Code)
	// Body
	fmt.Fprint(w, response.Body)

	// after
	log.Printf("request_out||client_ip=%s||uri=%s||code=%d||proc_time=%f\n", clientIp, r.RequestURI, response.Code, duration)

	// metric, 请求计数
	metric.HTTPReqTotal.With(prometheus.Labels{
		"method": r.Method,
		"path":   r.RequestURI,
		"status": strconv.Itoa(response.Code),
	}).Inc()
	// 请求处理时长
	metric.HTTPReqDuration.With(prometheus.Labels{
		"method": r.Method,
		"path":   r.RequestURI,
	}).Observe(duration)
}

// Server is httpserver
type Server struct {
	s   *http.Server
	mux *http.ServeMux
}

// Route 类似 HandleFunc
func (ser Server) Route(pattern string, handle httpHandle) {
	ser.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		httpHandleFunc(w, r, handle)
	})
}

// Run 启动 Server
func (ser Server) Run() error {
	return ser.s.ListenAndServe()
}

// New 返回一个新的 Server
// Server 应用 Mux 路由
func New() Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler()) // 支持 prom

	return Server{
		mux: mux,
		s:   &http.Server{Addr: ":8080", Handler: mux},
	}
}

func httpHandleHealthz(request *http.Request) Response {
	return Response{Body: "ok"}
}

func httpHandleHeaders(request *http.Request) Response {
	headers := request.Header.Clone()
	// set version
	var version string
	if version = os.Getenv("VERSION"); version == "" {
		// default 0.1
		version = "0.1"
	}
	headers.Set("version", version)
	return Response{Header: headers}
}

func main() {
	server := New()
	server.Route("/healthz", httpHandleHealthz)
	server.Route("/headers", httpHandleHeaders)

	log.Println("start http server")
	log.Fatal(server.Run())
}