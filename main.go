package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/l10n-center/api/src"
	"github.com/l10n-center/api/src/config"
	"github.com/l10n-center/api/src/errs"
	"github.com/l10n-center/api/src/store"
	"github.com/l10n-center/api/src/tracing"

	"github.com/tomazk/envcfg"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Default()

	if err := envcfg.Unmarshal(cfg); err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}

	l := initLogger(cfg)

	defer l.Sync()

	jCloser, err := tracing.Init(cfg)
	if err != nil {
		l.Fatal(err.Error(), errs.ZapStack(err))
	}

	defer jCloser.Close()

	store, err := store.New(cfg)
	if err != nil {
		l.Fatal(err.Error(), errs.ZapStack(err))
	}

	s := &http.Server{
		Addr:     cfg.Bind,
		Handler:  api.NewRouter(cfg, store),
		ErrorLog: zap.NewStdLog(zap.L()),
	}

	l.Info("starting http service")
	go s.ListenAndServe()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	l.Info("stopping http service")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
	}()
	if err := s.Shutdown(ctx); err != nil {
		l.Fatal(err.Error())
	}
	l.Info("stopped")
}

func initLogger(cfg *config.Config) *zap.Logger {
	lCfg := zap.NewProductionConfig()
	lCfg.DisableStacktrace = true // We use errs.ZapStack to get stacktrace
	if cfg.Debug {
		lCfg.Level.SetLevel(zap.DebugLevel)
	}
	l, err := lCfg.Build()
	if err != nil {
		log.Fatal(err.Error())
	}
	l = l.Named("l10n_center.api")
	zap.ReplaceGlobals(l)

	return l
}
