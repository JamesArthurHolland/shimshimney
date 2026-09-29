package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/shimshimney/pkg/api"
	pclient "github.com/shimshimney/pkg/client"
	"github.com/shimshimney/pkg/config"
	shlogger "github.com/shimshimney/pkg/logger"
	"github.com/shimshimney/shim/internal/runner"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfgPath := os.Getenv("SHIMNEY_CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.Config{
			BuildMode: "hot",
		}
	}
	cfg = config.ApplyEnv(cfg)
	logger := shlogger.New("shim")
	runner := runner.New(cfg, logger)
	if cfg.IsCold() {
		return runner.Rebuild()
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "backend"
	}
	podID := os.Getenv("POD_ID")
	if podID == "" {
		podID = fmt.Sprintf("%s-%d", name, rand.Intn(1000000))
	}
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		return fmt.Errorf("POD_NAMESPACE is required in hot mode")
	}
	operatorURL := cfg.OperatorURL
	if operatorURL == "" {
		operatorURL = "http://operator:8080"
	}
	client := pclient.New(operatorURL)
	host := os.Getenv("HOST")
	req := api.RegisterRequest{Namespace: namespace, PodID: podID, Name: name, Port: mustInt(port), Host: host}
	if err := client.Register(req); err != nil {
		logger.Warn("register failed", slog.String("error", err.Error()))
	}
	m := http.NewServeMux()
	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: "ok", Timestamp: time.Now()})
	})
	heartbeatReq := api.HeartbeatRequest{Namespace: namespace, PodID: podID, Name: name, Host: host, Port: mustInt(port), State: "healthy"}
	m.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if err := heartbeat(client, req, heartbeatReq, logger); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	m.HandleFunc("/rebuild", func(w http.ResponseWriter, r *http.Request) {
		if err := runner.Restart(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.MessageResponse{OK: true, Message: "rebuilt and restarted"})
	})
	if cfg.IsHot() {
		if err := runner.Rebuild(); err != nil {
			return err
		}
	}
	if err := runner.Start(); err != nil {
		return err
	}
	defer runner.Stop()
	go func() {
		for range time.Tick(10 * time.Second) {
			if err := heartbeat(client, req, heartbeatReq, logger); err != nil {
				logger.Warn("heartbeat failed", slog.String("error", err.Error()))
			}
		}
	}()
	log.Printf("shim serving on :%s", port)
	return http.ListenAndServe(":"+port, m)
}

// heartbeat sends a heartbeat and re-registers if the operator no longer knows
// this pod, which happens whenever the operator restarts and loses its
// in-memory registry.
func heartbeat(client *pclient.Client, reg api.RegisterRequest, hb api.HeartbeatRequest, logger *slog.Logger) error {
	err := client.Heartbeat(hb)
	if !pclient.IsNotFound(err) {
		return err
	}
	logger.Info("operator does not know this pod, re-registering")
	return client.Register(reg)
}

func mustInt(value string) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 8080
	}
	return v
}
