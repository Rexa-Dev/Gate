package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rexa-dev/Gate/config"
	"github.com/rexa-dev/Gate/controller"
	"github.com/rexa-dev/Gate/controller/rest"
	"github.com/rexa-dev/Gate/controller/rpc"
	"github.com/rexa-dev/Gate/tools"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.GateHost, cfg.ServicePort)

	tlsConfig, err := tools.LoadTLSCredentials(cfg.SslCertFile, cfg.SslKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting Gate: v%s", controller.GateVersion)

	var shutdownFunc func(ctx context.Context) error
	var service controller.Service

	if cfg.ServiceProtocol == "rest" {
		shutdownFunc, service, err = rest.StartHttpListener(tlsConfig, addr, cfg)
	} else {
		shutdownFunc, service, err = rpc.StartGRPCListener(tlsConfig, addr, cfg)
	}
	if err != nil {
		log.Fatal(err)
	}

	defer service.Disconnect()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt
	<-stopChan
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = shutdownFunc(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}
