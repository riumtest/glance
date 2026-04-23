package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Pages    []PageConfig   `yaml:"pages"`
	Branding BrandingConfig `yaml:"branding"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	AssetsPath string `yaml:"assets-path"`
}

// BrandingConfig holds customization options for the UI.
type BrandingConfig struct {
	CustomCSS    string `yaml:"custom-css"`
	FaviconURL   string `yaml:"favicon-url"`
	LogoURL      string `yaml:"logo-url"`
	SiteName     string `yaml:"site-name"`
}

// PageConfig represents a single dashboard page.
type PageConfig struct {
	Name    string        `yaml:"name"`
	Slug    string        `yaml:"slug"`
	Columns []ColumnConfig `yaml:"columns"`
}

// ColumnConfig represents a column within a page.
type ColumnConfig struct {
	Size    string        `yaml:"size"`
	Widgets []WidgetConfig `yaml:"widgets"`
}

// WidgetConfig holds the configuration for a single widget.
// The Type field determines which widget is rendered.
type WidgetConfig struct {
	Type     string            `yaml:"type"`
	Title    string            `yaml:"title"`
	Cache    time.Duration     `yaml:"cache"`
	Options  map[string]interface{} `yaml:"options,omitempty"`
}

// DefaultServerConfig returns sensible defaults for the server.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	}
}

// Load reads and parses a YAML configuration file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := &Config{
		Server: DefaultServerConfig(),
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// validate performs basic semantic validation on the loaded configuration.
func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port %d is out of range (1-65535)", c.Server.Port)
	}

	slugs := make(map[string]bool)
	for i, page := range c.Pages {
		if page.Name == "" {
			return fmt.Errorf("page at index %d is missing a name", i)
		}
		slug := page.Slug
		if slug == "" {
			slug = page.Name
		}
		if slugs[slug] {
			return fmt.Errorf("duplicate page slug %q", slug)
		}
		slugs[slug] = true
	}

	return nil
}

// Addr returns the formatted listen address for the HTTP server.
func (s *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
