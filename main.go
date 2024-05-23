package main

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// gauge metric
var onlineUsers = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "goapp_online_users",
	Help: "Online users",
	ConstLabels: map[string]string{
		"site": "ecommerce",
	},
})

// counter metric
var httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "goapp_http_requests_total",
	Help: "Count of all HTTP requests",
}, []string{})

// histogram metric
var httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "goapp_http_request_duration",
	Help: "Duration in seconds of all HTTP requests",
}, []string{"handler"})

func main() {
	r := prometheus.NewRegistry()

	r.MustRegister(onlineUsers)
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpDuration)

	go func() {
		for {
			onlineUsers.Set(float64(rand.Intn(2000)))
		}
	}()

	home := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello user"))
	})

	handlerForMeasureDurationAndCountHTTPRequests := promhttp.InstrumentHandlerDuration(
		httpDuration.MustCurryWith(prometheus.Labels{"handler": "home"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, home),
	)

	//http.Handle("/", promhttp.InstrumentHandlerCounter(httpRequestsTotal, home)) using the counter metric
	http.Handle("/", handlerForMeasureDurationAndCountHTTPRequests)
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	log.Fatal(http.ListenAndServe(":8181", nil))
}
