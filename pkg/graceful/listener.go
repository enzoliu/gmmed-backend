package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	DefaultConfig = ShutdownerConfig{
		ShutdownTimeout: 10 * time.Second,
	}
)

type ShutdownFunc func(ctx context.Context)

type ShutdownerConfig struct {
	ShutdownTimeout time.Duration
}

type Shutdowner struct {
	ShutdownerConfig
	deferredFunc []ShutdownFunc
	signalCtx    context.Context
	signalCancel context.CancelFunc
	finishCh     chan struct{}
}

func StartShutdowner(ctx context.Context) (*Shutdowner, context.Context) {
	return StartShutdownerWithConfig(ctx, DefaultConfig)
}

func StartShutdownerWithConfig(ctx context.Context, cfg ShutdownerConfig) (*Shutdowner, context.Context) {
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = DefaultConfig.ShutdownTimeout
	}

	signalCtx, signalCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	shutdowner := &Shutdowner{
		ShutdownerConfig: cfg,
		signalCtx:        signalCtx,
		signalCancel:     signalCancel,
		finishCh:         make(chan struct{}, 1),
	}

	go shutdowner.listen()

	return shutdowner, signalCtx
}

func (s *Shutdowner) DeferShutdownFunc(f ShutdownFunc) {
	s.deferredFunc = append(s.deferredFunc, f)
}

func (s *Shutdowner) Wait() {
	<-s.finishCh
}

func (s *Shutdowner) Close() {
	s.signalCancel()
}

func (s *Shutdowner) CloseAndWait() {
	s.Close()
	s.Wait()
}

func (s *Shutdowner) listen() {
	<-s.signalCtx.Done()

	// run shutdown functions
	shutDownCtx, shutDownCancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer shutDownCancel()

	for i := len(s.deferredFunc) - 1; i >= 0; i-- {
		s.deferredFunc[i](shutDownCtx)
	}

	close(s.finishCh)
}
