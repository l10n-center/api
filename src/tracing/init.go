package tracing

import (
	"io"

	"github.com/l10n-center/api/src/config"

	"github.com/pkg/errors"
	jConfig "github.com/uber/jaeger-client-go/config"
)

// Init global tracer
func Init(cfg *config.Config) (io.Closer, error) {
	jCfg := jConfig.Configuration{
		Reporter: &jConfig.ReporterConfig{
			LocalAgentHostPort: cfg.Jaeger,
		},
		Sampler: &jConfig.SamplerConfig{
			Type:  "const",
			Param: 1.0, // sample all traces
		},
	}
	closer, err := jCfg.InitGlobalTracer("l10n_center.api")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return closer, nil
}
