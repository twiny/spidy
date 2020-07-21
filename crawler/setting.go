package crawler

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Setting struct
type Setting struct {
	Engine *Engine `yaml:"Engine"`
}

// Engine struct
type Engine struct {
	Worker      int      `yaml:"worker"`
	Parallel    int      `yaml:"parallel"`
	Depth       int      `yaml:"depth"`
	URLPath     string   `yaml:"urls"`
	ProxyList   []string `yaml:"proxies,flow"`
	TLDs        []string `yaml:"tlds,flow"`
	Level       string   `yaml:"level"`
	RandomDelay string   `yaml:"random_delay"`
	Timeout     string   `yaml:"timeout"`
}

/*
SettingFromFile func
*/
func SettingFromFile(path string) (*Setting, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var setting Setting

	return &setting, yaml.Unmarshal(b, &setting)
}
