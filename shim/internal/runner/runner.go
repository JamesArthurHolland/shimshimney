package runner

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/shimshimney/pkg/config"
)

type Runner struct {
	config    config.Config
	logger    *slog.Logger
	BuildMode string
	BuildCmd  string
	RunCmd    string
	WorkDir   string
	AppPort   string

	mu      sync.Mutex
	current *exec.Cmd
	done    chan struct{}
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
	return &Runner{
		config:    cfg,
		logger:    logger,
		BuildMode: buildMode,
		BuildCmd:  buildCmd,
		RunCmd:    runCmd,
		WorkDir:   os.Getenv("APP_DIR"),
		AppPort:   os.Getenv("APP_PORT"),
	}
}

func (r *Runner) Rebuild() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.build()
}

func (r *Runner) build() error {
	if strings.TrimSpace(r.BuildCmd) == "" {
		return nil
	}
	cmd := exec.Command("sh", "-c", r.BuildCmd)
	r.configureCommand(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	r.logger.Info("project rebuild complete")
	return nil
}

// Start launches the run command in the background.
func (r *Runner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.start()
}

// Restart rebuilds the project and, only if the build succeeds, replaces the
// running process with a fresh one so the new binary is served.
func (r *Runner) Restart() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.build(); err != nil {
		return err
	}
	r.stop()
	return r.start()
}

// Stop terminates the running process, if any.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stop()
}

func (r *Runner) start() error {
	if strings.TrimSpace(r.RunCmd) == "" {
		return nil
	}
	cmd := exec.Command("sh", "-c", "exec "+r.RunCmd)
	r.configureCommand(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}
	done := make(chan struct{})
	r.current, r.done = cmd, done
	go func() {
		err := cmd.Wait()
		close(done)
		if err != nil {
			r.logger.Warn("project process exited", "error", err.Error())
		}
	}()
	r.logger.Info("project process started", "pid", cmd.Process.Pid)
	return nil
}

func (r *Runner) stop() {
	if r.current == nil {
		return
	}
	pgid := -r.current.Process.Pid
	_ = syscall.Kill(pgid, syscall.SIGTERM)
	select {
	case <-r.done:
	case <-time.After(10 * time.Second):
		_ = syscall.Kill(pgid, syscall.SIGKILL)
		<-r.done
	}
	r.current, r.done = nil, nil
}

func (r *Runner) configureCommand(cmd *exec.Cmd) {
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	if r.AppPort != "" {
		cmd.Env = append(os.Environ(), "PORT="+r.AppPort)
	}
}
