package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
)

func Instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)
		duration := float64(m.Duration / time.Millisecond)

		rctx := chi.RouteContext(r.Context())
		routePattern := strings.Join(rctx.RoutePatterns, "")
		routePattern = strings.Replace(routePattern, "/*/", "/", -1)

		requestDuration.WithLabelValues(
			r.Method, routePattern, strconv.Itoa(m.Code),
		).Observe(duration)

	})
}

func main() {
	n := chi.NewRouter()
	m := chiprometheus.NewMiddleware("serviceName")
	// if you want to use other buckets than the default (300, 1200, 5000) you can run:
	// m := negroniprometheus.NewMiddleware("serviceName", 400, 1600, 700)

	n.Use(m)

	n.Handle("/metrics", prometheus.Handler())
	n.Get("/ok", func(w http.ResponseWriter, r *http.Request) {
		sleep := rand.Intn(4999) + 1
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "slept %d milliseconds\n", sleep)
	})
	n.Get("/notfound", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "not found")
	})

	http.ListenAndServe(":3000", n)
}
