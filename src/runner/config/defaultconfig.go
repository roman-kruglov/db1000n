package config

import (
	"encoding/base64"
)

// DefaultConfig is the config embedded into the app that it will use if not able to fetch any other config
var DefaultConfig = `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`

func init() {
	decoded, err := base64.StdEncoding.DecodeString(DefaultConfig)
	if err != nil {
		panic("Can't decode base64 encoded encrypted config")
	}

	DefaultConfig = string(decoded)
}
