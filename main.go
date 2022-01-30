package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

var RequestCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "go_app_requests_count",
	Help: "Total App Http Request Count",
})

var RequestInprogress = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "go_app_requests_inprogress",
	Help: "Number of application requests in progress",
})

var RequestResponseTime = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "go_app_response_latency_seconds",
	Help: "Response latency in seconds",
}, []string{"path"})

var RequestResponseTimeByHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "go_app_response_latency_seconds_histogram",
	Help: "Response latency in seconds by Histogram",
}, []string{"path"})

func routeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		next.ServeHTTP(w, r)
		timeTaken := time.Since(startTime)
		RequestResponseTime.WithLabelValues(path).Observe(timeTaken.Seconds())
		RequestResponseTimeByHistogram.WithLabelValues(path).Observe(timeTaken.Seconds())
	})
}

func main(){
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/user/{name}", func(rw http.ResponseWriter, r *http.Request) {
		RequestInprogress.Inc() // arttır
		vars := mux.Vars(r)
		name := vars["name"]
		greetings := fmt.Sprintf("Username %s :)", name)
		time.Sleep(6 * time.Second)
		_, err := rw.Write([]byte(greetings))
		if err != nil {
			return
		}
		RequestCount.Inc()
	}).Methods("GET")

	router.Use(routeMiddleware)
	log.Println("Starting the application server")
	router.Path("/metrics").Handler(promhttp.Handler())
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		return
	}
}