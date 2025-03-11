package httpretry

import (
	"log"
	"time"

	"github.com/Netflix/go-env"
)

// Settings of http retry
type (
	Settings struct {
		MaxRetry              int           `env:"MAX_REQUEST_RETRY,default=3"`
		DebugMode             bool          `env:"DEBUG_MODE,default=false"`
		Insecure              bool          `env:"INSECURE,default=false"`
		MaxIdleConns          int           `env:"MAX_IDLE_CONNECTIONS,default=15"`
		IdleConnTimeout       time.Duration `env:"CONNECTION_TIMEOUT,default=90s"`
		TLSHandshakeTimeout   time.Duration `env:"TLS_TIMEOUT,default=10s"`
		ExpectContinueTimeout time.Duration `env:"CONTINUE_TIMEOUT,defualt=1s"`
		ResponseHeaderTimeout time.Duration `env:"HEADER_TIMEOUT,default=10s"`
		RequestTimeout        time.Duration `env:"REQUEST_TIMEOUT,default=10s"`
		BackoffPolicy         func(attempt int) time.Duration
	}
)

// NewSettings constructor
func NewSettings() *Settings {
	var settings Settings
	_, err := env.UnmarshalFromEnviron(&settings)
	if err != nil {
		log.Fatal(err)
	}
	return &settings
}
