package runner

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/shimshimney/pkg/config"
)

type Runner struct {
	config     config.Config
	logger     *slog.Logger
	BuildMode  string
	BuildCmd   string
	RunCmd     string
}

func New(cfg config.Config, logger *slog.Logger) *Runner {
	buildMode := cfg.BuildMode
	if buildMode == "" {
		buildMode = "hot"
	}
	buildCmd := cfg.Build
	if buildCmd == "" {
		buildCmd = "go build ./..."
	}
	runCmd := cfg.Run
	if runCmd == "" {
		runCmd = "go run ."
	}
	return &Runner{config: cfg, logger: logger, BuildMode: buildMode, BuildCmd: buildCmd, RunCmd: runCmd}
}

func (r *Runner) Rebuild() error {
	if strings.TrimSpace(r.BuildCmd) == "" {
		return nil
	}
	cmd := exec.Command("sh", "-c", r.BuildCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	r.logger.Info("project rebuild complete")
	return nil
}

func (r *Runner) Run() error {
	if strings.TrimSpace(r.RunCmd) == "" {
		return nil
	}
	cmd := exec.Command("sh", "-c", r.RunCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
