package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dsociative/best_scraper/service_response"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	addr      = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	sitesPath = flag.String("sites", "./sites.txt", "path to sites.txt")
)

var (
	serviceRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "best_scraper",
			Name:      "service_request_count",
			Help:      "request count by site name",
		},
		[]string{"site"},
	)
)

func init() {
	prometheus.MustRegister(serviceRequestCount)
}

func makeEndpoint(f func() (service_response.ResponseResult, error)) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		r, err := f()
		if err == nil {
			serviceRequestCount.WithLabelValues(r.Site).Inc()
		}
		return r, err
	}
}

func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func heartbeat(sites []string, siteChan chan string) {
	for {
		nextTick := time.After(time.Minute)
		for _, s := range sites {
			siteChan <- s
		}
		_ = <-nextTick
	}
}

func main() {
	flag.Parse()

	raw, err := ioutil.ReadFile(*sitesPath)
	if err != nil {
		log.Fatal(err)
	}
	sites := strings.Split(string(raw), "\n")

	siteChan := make(chan string)
	resultChan := make(chan service_response.ResponseResult, len(sites))
	store := service_response.NewResponseTimeStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < runtime.NumCPU(); i++ {
		go service_response.ServiceResponseTimeWorker(ctx, siteChan, resultChan)
	}
	go store.Listen(resultChan)
	go heartbeat(sites, siteChan)

	minHandler := httptransport.NewServer(
		makeEndpoint(func() (service_response.ResponseResult, error) {
			return store.Min()
		}),
		decodeRequest,
		encodeResponse,
	)
	maxHandler := httptransport.NewServer(
		makeEndpoint(func() (service_response.ResponseResult, error) {
			return store.Max()
		}),
		decodeRequest,
		encodeResponse,
	)
	randomHandler := httptransport.NewServer(
		makeEndpoint(func() (service_response.ResponseResult, error) {
			return store.Random()
		}),
		decodeRequest,
		encodeResponse,
	)

	http.Handle("/min", minHandler)
	http.Handle("/max", maxHandler)
	http.Handle("/random", randomHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(*addr, nil))
}
