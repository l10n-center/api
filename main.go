package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/l10n-center/api/src/controller"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
	jaegerClientConfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/source/file"
)

func main() {
	l := initLogger()

	defer l.Sync()

	jCloser := initTracing(l)

	defer jCloser.Close()

	db := initDB(l)

	defer db.Close()

	s := initServer(l, db)

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

func initLogger() *zap.Logger {
	var (
		l   *zap.Logger
		err error
	)
	if len(os.Getenv("PRODUCTION")) > 0 {
		l, err = zap.NewProduction()
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(l)

	return l
}

func initTracing(l *zap.Logger) io.Closer {
	JAEGER := os.Getenv("JAEGER")
	if len(JAEGER) == 0 {
		JAEGER = "localhost:5775"
		l.Sugar().Infof("jaeger is not set, use default (%s)", JAEGER)
	}

	jcfg := jaegerClientConfig.Configuration{
		Reporter: &jaegerClientConfig.ReporterConfig{
			LocalAgentHostPort: JAEGER,
		},
		Sampler: &jaegerClientConfig.SamplerConfig{
			Type:  "const",
			Param: 1.0, // sample all traces
		},
	}
	closer, err := jcfg.InitGlobalTracer("l10n-center/api")
	if err != nil {
		l.Fatal(err.Error())
	}

	return closer
}

func initDB(l *zap.Logger) *sql.DB {
	DBURL := os.Getenv("DB")
	if len(DBURL) == 0 {
		DBURL = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
		l.Sugar().Warnf("db is not set, use default (%s)", DBURL)
	}
	l.Info("connecting to db")
	db, err := sql.Open("postgres", DBURL)
	if err != nil {
		l.Fatal(err.Error())
	}

	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)

	l.Info("migrating")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		l.Fatal(err.Error())
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://./migration",
		"postgres", driver)
	if err != nil {
		l.Fatal(err.Error())
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			l.Info(err.Error())
		} else {
			l.Fatal(err.Error())
		}
	}

	return db
}

func initServer(l *zap.Logger, db *sql.DB) *http.Server {
	SECRET := os.Getenv("SECRET")
	if len(SECRET) == 0 {
		l.Warn("secret is not set, generating random")
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		buf := make([]byte, 20)
		rnd.Read(buf)
		SECRET = base64.URLEncoding.EncodeToString(buf)
	}
	BIND := os.Getenv("BIND")
	if len(BIND) == 0 {
		BIND = "0.0.0.0:3000"
		l.Sugar().Infof("bind not set, use default (%s)", BIND)
	}
	r := controller.NewRouter(&controller.Config{
		Secret: []byte(SECRET),
	}, db)

	return &http.Server{
		Addr:     BIND,
		Handler:  r,
		ErrorLog: zap.NewStdLog(zap.L()),
	}
}
