package spider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type (
	Config struct {
		HomeDir string `json:"home_dir"`
		Logs    struct {
			MaxAge  int    `json:"max_age"`
			MaxSize int    `json:"max_size"`
			Path    string `json:"path"`
		} `json:"logs"`
		Store struct {
			TTL  time.Duration `json:"ttl"`
			Path string        `json:"path"`
		} `json:"store"`
		Crawler CrawlerConfig `json:"crawler"`
	}
	CrawlerConfig struct {
		Worker      int           `json:"worker"`
		MaxDepth    int32         `json:"max_depth"`
		Filter      []string      `json:"filter"`
		RateLimit   string        `json:"rate_limit"`
		MaxBodySize string        `json:"max_body_size"`
		Timeout     time.Duration `json:"timeout"`
		UserAgents  []string      `json:"user_agents"`
		Proxies     []string      `json:"proxies"`
		AllowedTLDs []string      `json:"allowed_tlds"`
	}
)

var (
	core = func() int {
		c := runtime.NumCPU()
		if c == 1 {
			return c
		}
		return c - 1
	}()
)

func InitConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	spidyDir := filepath.Join(homeDir, ".spidy")

	if err := os.MkdirAll(spidyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write initial config.json file inside ~/.spidy/.
	config := &Config{
		HomeDir: homeDir,
		Logs: struct {
			MaxAge  int    `json:"max_age"`
			MaxSize int    `json:"max_size"`
			Path    string `json:"path"`
		}{
			MaxAge:  7,
			MaxSize: 100,
			Path:    filepath.Join(spidyDir, "log", "spidy.log"),
		},
		Store: struct {
			TTL  time.Duration `json:"ttl"`
			Path string        `json:"path"`
		}{
			TTL:  24 * time.Hour,
			Path: filepath.Join(spidyDir, "storage"),
		},
		Crawler: CrawlerConfig{
			Worker:      core,
			MaxDepth:    10,
			Filter:      []string{},
			RateLimit:   "10/s",
			MaxBodySize: "5mb",
			Timeout:     60 * time.Second,
			UserAgents: []string{
				"Spidy/2.1; +",
			},
			Proxies: []string{
				"localhost",
			},
		},
	}

	configFile, err := os.Create(filepath.Join(spidyDir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to create config.json: %w", err)
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to encode config.json: %w", err)
	}

	// 4. Create the ~/.spidy/storage and ~/.spidy/log directories.
	for _, dir := range []string{"storage", "log"} {
		if err := os.Mkdir(filepath.Join(spidyDir, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return config, nil
}

func GetConfig() (*Config, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine user's home directory: %w", err)
	}

	configFile, err := os.Open(filepath.Join(dir, ".spidy", "config.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	return &config, nil
}

func ParseCrawlerConfig(file string) (*CrawlerConfig, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer configFile.Close()

	var config CrawlerConfig
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	return &config, nil
}
