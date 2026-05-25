package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pokgak/lustrefs-exporter/collector"
)

func main() {
	port := flag.String("port", "32221", "Listen port")
	lustrePath := flag.String("lustre-path", "/proc/fs/lustre", "Path to Lustre procfs")
	lnetctlBin := flag.String("lnetctl", "lnetctl", "Path to lnetctl binary")
	flag.Parse()

	if _, err := os.Stat(*lustrePath); os.IsNotExist(err) {
		log.Printf("WARNING: lustre-path %q does not exist; serving empty metrics", *lustrePath)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.New(*lustrePath, *lnetctlBin))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><a href="/metrics">metrics</a></body></html>`)) //nolint
	})

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	go func() {
		log.Printf("Listening on :%s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
