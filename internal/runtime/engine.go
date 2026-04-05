package runtime

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christianmscott/overwatch/internal/alerts"
	"github.com/christianmscott/overwatch/internal/api"
	"github.com/christianmscott/overwatch/internal/config"
	"github.com/christianmscott/overwatch/internal/results"
	"github.com/christianmscott/overwatch/internal/scheduler"
	"github.com/christianmscott/overwatch/internal/version"
	"github.com/christianmscott/overwatch/internal/worker"
	"github.com/christianmscott/overwatch/pkg/spec"
)

type Engine struct {
	cfg     *spec.Config
	cfgPath string
}

func NewEngine(cfg *spec.Config, cfgPath string) *Engine {
	return &Engine{cfg: cfg, cfgPath: cfgPath}
}

func (e *Engine) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := results.NewStore(100)
	srv := api.New(e.cfg, e.cfgPath, store)

	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go func() {
		for range sighup {
			slog.Info("SIGHUP received, reloading config", "path", e.cfgPath)
			newCfg, err := config.Load(e.cfgPath)
			if err != nil {
				slog.Error("config reload failed", "error", err)
				continue
			}
			e.cfg = newCfg
			srv.UpdateConfig(newCfg)
			slog.Info("config reloaded", "checks", len(newCfg.Checks))
		}
	}()

	go func() {
		if err := srv.Serve(ctx); err != nil {
			slog.Error("api server error", "error", err)
		}
	}()

	source := NewLocalJobSource(e.cfg.Checks)

	wi := spec.WorkerInfo{
		ID:      hostname(),
		Version: version.Version,
	}

	tick := 1 * time.Second
	sched := scheduler.New(source, wi, tick, len(e.cfg.Checks)*2+8)

	senders := alerts.BuildSenders(e.cfg.Alerts)
	router := alerts.NewRouter(senders)

	if len(senders) > 0 {
		slog.Info("alerting enabled", "senders", len(senders))
	}

	handleResult := func(r spec.CheckResult) {
		store.Record(r)

		attrs := []any{
			"check", r.CheckName,
			"status", r.Status,
			"duration", r.Duration,
		}
		if r.Error != "" {
			attrs = append(attrs, "error", r.Error)
		}
		switch r.Status {
		case spec.StatusUp:
			slog.Info("check complete", attrs...)
		case spec.StatusDegraded:
			slog.Warn("check complete", attrs...)
		default:
			slog.Error("check complete", attrs...)
		}

		router.Handle(r)
	}

	pool := worker.NewPool(e.cfg.Server.Concurrency, source, handleResult)

	slog.Info("starting overwatch",
		"checks", len(e.cfg.Checks),
		"concurrency", e.cfg.Server.Concurrency,
		"api", srv.Addr(),
		"version", version.Version,
	)

	go sched.Run(ctx)
	pool.Run(ctx, sched.C())

	slog.Info("shutting down")
	return nil
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
