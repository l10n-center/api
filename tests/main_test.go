package tests

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
	jaegerClientConfig "github.com/uber/jaeger-client-go/config"

	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/source/file"
)

var (
	DBURL  string
	maindb *sql.DB
)

func TestMain(m *testing.M) {
	var err error

	DBURL = os.Getenv("DB")
	if len(DBURL) == 0 {
		DBURL = "postgres://postgres@localhost:5432/%s?sslmode=disable"
	}
	maindb, err = sql.Open("postgres", fmt.Sprintf(DBURL, "postgres"))
	if err != nil {
		panic(err)
	}

	jcloser := initTracing()

	st := m.Run()

	jcloser.Close()
	maindb.Close()

	os.Exit(st)
}

func initTracing() io.Closer {
	JAEGER := os.Getenv("JAEGER")
	if len(JAEGER) == 0 {
		JAEGER = "localhost:5775"
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
		panic(err)
	}

	return closer
}

func initDB(name string) *sql.DB {
	maindb.Exec(fmt.Sprintf("DROP DATABASE %q", name))
	_, err := maindb.Exec(fmt.Sprintf("CREATE DATABASE %q", name))
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", fmt.Sprintf(DBURL, name))
	if err != nil {
		panic(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../migration",
		"postgres", driver)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			panic(err.Error())
		}
	}

	return db
}

func readBody(res *http.Response) []byte {
	buf := &bytes.Buffer{}
	io.Copy(buf, res.Body)

	return buf.Bytes()
}

func printUberTraceID(t *testing.T, res *http.Response) {
	t.Logf("Uber-Trace-Id: %s", strings.SplitN(res.Header.Get("Uber-Trace-Id"), ":", 2)[0])
}
