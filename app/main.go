package main

import (
	"flag"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	maxImageSize   = 1024 * 1000
	maxImageWidth  = 1000
	maxImageHeight = 1000
	cacheLimitSize = 1024 * 1024 * 50
	cacheLiveTime  = time.Second * 60 * 60
)

func main() {
	log.Printf("[INFO] started")

	listenPort := flag.Int("p", 8189, "listen port")
	flag.Parse()

	pctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	cacheService = newCache(cacheLimitSize, cacheLiveTime)

	router := chi.NewRouter()
	router.Use(middleware.Throttle(500), middleware.Timeout(time.Second*60))
	router.Get("/", handler)

	addr := fmt.Sprintf(":%d", *listenPort)
	srv := http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Second * 30,
	}

	go func() {
		log.Printf("[INFO] start listen on %s %+v", addr, os.Args)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			log.Printf("[ERROR] failed to start web server %v", err)
			cancel()
		}
	}()

	select {
	case v := <-sigChan:
		log.Printf("[INFO] received signal %v", v)
	case <-pctx.Done():
	}

	log.Printf("[INFO] start shutdown")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] shutdown error, %v", err)
	}

	log.Printf("[INFO] exited")
}
