package spider

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"

	//
	"github.com/twiny/spidy/v2/internal/pkg/hbyte"

	"gopkg.in/yaml.v3"
)

// default cores
var core = func() int {
	c := runtime.NumCPU()
	if c == 1 {
		return c
	}
	return c - 1
}()

// defaultSetting
var defaultSetting = &Setting{
	Crawler: struct {
		MaxDepth int32
		Filter   []string
		Limit    struct {
			Rate     int
			Interval time.Duration
		}
		MaxBodySize int64
		UserAgents  []string
		Proxies     []string
	}{
		MaxDepth: 10,
		Filter:   []string{},
		Limit: struct {
			Rate     int
			Interval time.Duration
		}{
			Rate:     1,
			Interval: time.Second,
		},
		MaxBodySize: 10 * 1024 * 1024, // 10 MB
		UserAgents:  []string{`Spidy/2.1; +https://github.com/twiny/spidy`},
		Proxies:     []string{},
	},
	Log: struct {
		Rotate int
		Path   string
	}{
		Rotate: 7,
		Path:   "./log",
	},
	Store: struct {
		TTL  time.Duration
		Path string
	}{
		TTL:  6 * time.Hour, // format: 1h, 1d, 1w, 1m - minimum 6h
		Path: "./store",
	},
	Result: struct{ Path string }{
		Path: "./result",
	},
	Parralle: core,
	Timeout:  1 * time.Minute,
	TLDs:     tlds,
}

// Setting
type Setting struct {
	Crawler struct {
		MaxDepth int32
		Filter   []string
		Limit    struct {
			Rate     int
			Interval time.Duration
		}
		MaxBodySize int64
		UserAgents  []string
		Proxies     []string
	}
	Log struct {
		Rotate int // format: 30d
		Path   string
	}
	Store struct {
		TTL  time.Duration
		Path string
	}
	Result struct {
		Path string
	}
	Parralle int
	Timeout  time.Duration
	TLDs     map[string]bool
}

// ParseSetting
func ParseSetting(fp string) *Setting {
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return defaultSetting
	}

	var s = struct {
		Crawler struct {
			MaxDepth    int32    `yaml:"max_depth"`
			Filter      []string `yaml:"filter,flow"`
			RateLimit   string   `yaml:"rate_limit"` // format: req/time.Duration => 5/1s
			MaxBodySize string   `yaml:"max_body_size"`
			UserAgents  []string `yaml:"user_agents,flow"`
			Proxies     []string `yaml:"proxies,flow"`
		} `yaml:"crawler"`
		Log struct {
			Rotate int    `yaml:"rotate"` // format: 30d
			Path   string `yaml:"path"`
		} `yaml:"log"`
		Store struct {
			TTL  string `yaml:"ttl"` // format: 1h, 24h
			Path string `yaml:"path"`
		} `yaml:"store"`
		Result struct {
			Path string `yaml:"path"`
		} `yaml:"result"`
		Parralle int      `yaml:"parralle"`
		Timeout  string   `yaml:"timeout"`
		TLDs     []string `yaml:"tlds,flow"`
	}{}

	if err := yaml.Unmarshal(data, &s); err != nil {
		return defaultSetting
	}

	rate, interval := parseRateLimit(s.Crawler.RateLimit)

	return &Setting{
		Crawler: struct {
			MaxDepth int32
			Filter   []string
			Limit    struct {
				Rate     int
				Interval time.Duration
			}
			MaxBodySize int64
			UserAgents  []string
			Proxies     []string
		}{
			MaxDepth: s.Crawler.MaxDepth,
			Filter:   s.Crawler.Filter,
			Limit: struct {
				Rate     int
				Interval time.Duration
			}{
				Rate:     rate,
				Interval: interval,
			},
			MaxBodySize: parseBodySize(s.Crawler.MaxBodySize),
			UserAgents:  s.Crawler.UserAgents,
			Proxies:     s.Crawler.Proxies,
		},
		Log: struct {
			Rotate int
			Path   string
		}{
			Rotate: s.Log.Rotate,
			Path:   s.Log.Path,
		},
		Store: struct {
			TTL  time.Duration
			Path string
		}{
			TTL:  parseTTL(s.Store.TTL),
			Path: s.Store.Path,
		},
		Result: struct{ Path string }{
			Path: s.Result.Path,
		},
		Parralle: s.Parralle,
		Timeout:  parseTimeout(s.Timeout),
		TLDs:     parseTLDs(s.TLDs),
	}
}

// parseRateLimit
func parseRateLimit(s string) (int, time.Duration) {
	// default rate limit
	dr, di := defaultSetting.Crawler.Limit.Rate, defaultSetting.Crawler.Limit.Interval

	if s == "" {
		return dr, di
	}

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return dr, di
	}

	r, i := parts[0], parts[1]

	rate, err := strconv.Atoi(r)
	if err != nil {
		return dr, di
	}

	interval, err := time.ParseDuration(i)
	if err != nil {
		return dr, di
	}

	return rate, interval
}

// parseTLDs
func parseTLDs(list []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range list {
		m[s] = true
	}
	return m
}

// parseTimeout
func parseTimeout(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultSetting.Timeout
	}
	return d
}

// parseTTL
func parseTTL(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultSetting.Timeout
	}
	return d
}

// parseBodySize
func parseBodySize(s string) int64 {
	size := hbyte.Parse(s)
	if size == 0 {
		return defaultSetting.Crawler.MaxBodySize
	}
	return size
}
