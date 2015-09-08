package uweb

import (
	"sync"
	
	"github.com/robfig/config"
)

//
// Config interface
//
type Config interface {
	Add(section, key, value string)
	Str(section, key string) string
	Int(section, key string) int
}

//
// Create config middleware
//
func MdConfig(file string) Middleware {
	cfg, err := NewIniCfg(file)
	if err != nil {
		panic(err)
	}
	return cfg
}

//
// Ini config
//
type IniCfg struct {
	mu   sync.RWMutex
	data *config.Config
}

func NewIniCfg(file string) (*IniCfg, error) {
	d, err := config.ReadDefault(file)
	if err != nil {
		return nil, err
	}
	return &IniCfg{
		data: d,
	}, nil
}

// @impl Middleware
func (cfg *IniCfg) Handle(c *Context) int {
	c.Cfg = cfg
	return NEXT_CONTINUE
}

// Add section and key-value
func (cfg *IniCfg) Add(section, key, value string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if !cfg.data.HasSection(section) {
		cfg.data.AddSection(section)
	}
	cfg.data.AddOption(section, key, value)
}

// Get string
func (cfg *IniCfg) Str(section, key string) string {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	value, err := cfg.data.String(section, key)
	if err != nil {
		panic(err)
	}
	return value
}

// Get int
func (cfg *IniCfg) Int(section, key string) int {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	value, err := cfg.data.Int(section, key)
	if err != nil {
		panic(err)
	}
	return value
}
