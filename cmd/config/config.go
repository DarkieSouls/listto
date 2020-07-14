package config

import (
	"os"
	"strings"

	"github.com/DarkieSouls/listto/internal/listtoErr"
)

// Config contains the configuration of the bot.
type Config struct {
	token  string
	prefix string
}

// NewConfig generates a new configuration based on current envvars.
func NewConfig() (c *Config, lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("NewConfig")
		}
	}()

	token := strings.TrimSpace(os.Getenv("LISTTO_TOKEN"))
	if token == "" {
		lisErr = listtoErr.InvalidEnvvar("token")
		return
	}

	prefix := strings.TrimSpace(os.Getenv("LISTTO_PREFIX"))
	if prefix == "" {
		prefix = "^"
	}

	c = new(Config)
	c.token = token
	c.prefix = prefix

	return
}

// SetPrefix to the specified value.
// Will not update if new value is nil or same as the previous one.
func (c *Config) SetPrefix(prefix string) {
	if strings.TrimSpace(prefix) == "" || strings.TrimSpace(prefix) == c.prefix {
		return
	}

	c.prefix = prefix
}

// Prefix returns the prefix stored in the config.
func (c *Config) Prefix() string {
	return c.prefix
}

// Token returns the token stored in the config.
func (c *Config) Token() string {
	return c.token
}
