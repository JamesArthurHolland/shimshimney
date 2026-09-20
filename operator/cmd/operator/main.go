package main

import (
	"fmt"
	"log"
	"os"

	shlogger "github.com/shimshimney/pkg/logger"
	"github.com/shimshimney/operator/internal/k8s"
	"github.com/shimshimney/operator/internal/registry"
	"github.com/shimshimney/operator/internal/server"
)

func main() {
	logger := shlogger.New("operator")
	reg := registry.New()
	services := k8s.NewManager()
	srv := server.NewServer(reg, services, logger)
	addr := os.Getenv("OPERATOR_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}
	logger.Info(fmt.Sprintf("operator listening on %s", addr))
	if err := srv.Run(addr); err != nil {
		log.Fatal(err)
	}
}
