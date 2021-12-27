package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//HTTPReqDuration metric:http_request_duration_seconds
	HTTPReqDuration *prometheus.HistogramVec
	//HTTPReqTotal metric:http_request_total
	HTTPReqTotal *prometheus.CounterVec
)

func init() {
	// 监控接口请求耗时
	// HistogramVec 是一组Histogram
	HTTPReqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "The HTTP request latencies in seconds.",
		Buckets: nil,
	}, []string{"method", "path"})
	// 这里的"method"、"path" 都是label
	// 监控接口请求次数
	// HistogramVec 是一组Histogram
	HTTPReqTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests made.",
	}, []string{"method", "path", "status"})
	// 这里的"method"、"path"、"status" 都是label
	prometheus.MustRegister(
		HTTPReqDuration,
		HTTPReqTotal,
	)
}