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
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.Config{
			OperatorURL: os.Getenv("OPERATOR_URL"),
			BuildMode:   os.Getenv("SHIMNEY_MODE"),
			Build:       os.Getenv("BUILD_COMMAND"),
			Run:         os.Getenv("RUN_COMMAND"),
		}
	}
	if cfg.OperatorURL == "" {
		cfg.OperatorURL = os.Getenv("OPERATOR_URL")
	}
	if cfg.BuildMode == "" {
		cfg.BuildMode = os.Getenv("SHIMNEY_MODE")
	}
	if cfg.Build == "" {
		cfg.Build = os.Getenv("BUILD_COMMAND")
	}
	if cfg.Run == "" {
		cfg.Run = os.Getenv("RUN_COMMAND")
	}
	if cfg.BuildMode == "" {
		cfg.BuildMode = "hot"
	}
	logger := shlogger.New("shim")
	runner := runner.New(cfg, logger)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "backend"
	}
	podID := os.Getenv("POD_ID")
	if podID == "" {
		podID = fmt.Sprintf("%s-%d", name, rand.Intn(1000000))
	}
	operatorURL := cfg.OperatorURL
	if operatorURL == "" {
		operatorURL = "http://operator:8080"
	}
	client := pclient.New(operatorURL)
	req := api.RegisterRequest{PodID: podID, Name: name, Port: mustInt(port), Host: os.Getenv("HOST"), BuildMode: cfg.BuildMode}
	if err := client.Register(req); err != nil {
		logger.Warn("register failed", slog.String("error", err.Error()))
	}
	m := http.NewServeMux()
	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: "ok", Timestamp: time.Now()})
	})
	m.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if err := client.Heartbeat(api.HeartbeatRequest{PodID: podID, Name: name, Port: mustInt(port), State: "healthy"}); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	m.HandleFunc("/rebuild", func(w http.ResponseWriter, r *http.Request) {
		if err := runner.Rebuild(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(api.MessageResponse{OK: true, Message: "rebuild initiated"})
	})
	go func() {
		for range time.Tick(10 * time.Second) {
			if err := client.Heartbeat(api.HeartbeatRequest{PodID: podID, Name: name, Port: mustInt(port), State: "healthy"}); err != nil {
				logger.Warn("heartbeat failed", slog.String("error", err.Error()))
			}
		}
	}()
	log.Printf("shim serving on :%s", port)
	if err := http.ListenAndServe(":"+port, m); err != nil {
		log.Fatal(err)
	}
}

func mustInt(value string) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return 8080
	}
	return v
}
