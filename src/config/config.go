package config

import (
	"crypto/rand"
	"encoding/base64"
)

// Config of application
type Config struct {
	Debug bool `envcfg:"L10N_SECRET"`
	// Secret is a random bytes to sign jwt
	Secret string `envcfg:"L10NC_SECRET"`
	// Bind string, default ":3000"
	Bind string `envcfg:"L10NC_BIND"`
	// Jaeger addr, default "localhost:5775"
	Jaeger string `envcfg:"L10NC_JAEGER"`
	// MongoHost addr, default "localhost:27017"
	MongoHost string `envcfg:"L10NC_MONGO_HOST"`
	// MongoDB name, default "l10n_center"
	MongoDB string `envcfg:"L10NC_MONGO_DB"`
}

// Default Config
func Default() *Config {
	c := &Config{
		Bind:      ":3000",
		Jaeger:    "localhost:5775",
		MongoHost: "localhost:27017",
		MongoDB:   "l10n_center",
	}

	buf := make([]byte, 15)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	c.Secret = base64.URLEncoding.EncodeToString(buf)

	return c
}
