// Package proc supervises a proxy-core child process for any engine. It keeps
// the process alive (restarting with backoff on unexpected exit) and exposes a
// small Start/Restart/Stop/Running surface the drivers build on. The exact
// command line differs per engine (xray uses "-config", sing-box uses "-c"), so
// the args are supplied by the caller.
package proc

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

// Supervisor owns the lifecycle of one core process.
type Supervisor struct {
	binPath string
	args    []string
	log     *slog.Logger

	mu      sync.Mutex
	cmd     *exec.Cmd
	running bool
	stopped bool // set by Stop; suppresses auto-restart

	logs *lineRing // recent core stdout/stderr lines
}

// New builds a supervisor that runs `binPath args...`.
func New(binPath string, args []string, log *slog.Logger) *Supervisor {
	if log == nil {
		log = slog.Default()
	}
	return &Supervisor{binPath: binPath, args: args, log: log, logs: newLineRing(1000)}
}

// Logs returns up to limit of the most recent captured core log lines (oldest
// first); limit <= 0 returns all retained lines.
func (s *Supervisor) Logs(limit int) []string { return s.logs.tail(limit) }

// Start launches the process if not already running. Idempotent.
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}
	s.stopped = false
	return s.spawnLocked(ctx)
}

// Restart replaces the running process with a fresh one (used to apply a new
// config for engines that cannot hot-reload).
func (s *Supervisor) Restart(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.killLocked()
	return s.spawnLocked(ctx)
}

// Stop terminates the process and disables auto-restart.
func (s *Supervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	s.killLocked()
}

// Running reports whether the child is currently up.
func (s *Supervisor) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Supervisor) spawnLocked(ctx context.Context) error {
	if s.binPath == "" {
		return errors.New("proc: empty binary path")
	}
	cmd := exec.Command(s.binPath, s.args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.cmd = cmd
	s.running = true
	go s.pumpLogs(stdout)
	go s.pumpLogs(stderr)
	go s.waitAndMaybeRestart(ctx, cmd)
	return nil
}

func (s *Supervisor) killLocked() {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	s.running = false
}

// waitAndMaybeRestart reaps the process and, unless Stop was called, restarts it
// after a short backoff so a crashing core self-heals.
func (s *Supervisor) waitAndMaybeRestart(ctx context.Context, cmd *exec.Cmd) {
	err := cmd.Wait()

	s.mu.Lock()
	s.running = false
	stopped := s.stopped
	s.mu.Unlock()

	if stopped || ctx.Err() != nil {
		return
	}
	s.log.Warn("core exited unexpectedly, restarting", "err", err)

	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Second):
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	if err := s.spawnLocked(ctx); err != nil {
		s.log.Error("core restart failed", "err", err)
	}
}

func (s *Supervisor) pumpLogs(r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		s.logs.add(line)
		s.log.Debug("core", "line", line)
	}
}

// lineRing is a fixed-capacity ring buffer of recent log lines, safe for
// concurrent writes (pump goroutines) and reads (the Logs RPC).
type lineRing struct {
	mu   sync.Mutex
	buf  []string
	next int
	full bool
}

func newLineRing(capacity int) *lineRing {
	if capacity <= 0 {
		capacity = 1000
	}
	return &lineRing{buf: make([]string, capacity)}
}

func (r *lineRing) add(line string) {
	r.mu.Lock()
	r.buf[r.next] = line
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
	r.mu.Unlock()
}

func (r *lineRing) tail(limit int) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ordered []string
	if r.full {
		ordered = append(ordered, r.buf[r.next:]...)
		ordered = append(ordered, r.buf[:r.next]...)
	} else {
		ordered = append(ordered, r.buf[:r.next]...)
	}
	if limit > 0 && len(ordered) > limit {
		ordered = ordered[len(ordered)-limit:]
	}
	return ordered
}
