package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type config struct {
	envPrefix  string
	flagSet    *flag.FlagSet
	values     map[string]interface{}
	alreadySet map[string]struct{}
	mtx        sync.RWMutex
}

func newConfig(flagSet *flag.FlagSet, envPrefix string) *config {
	return &config{
		envPrefix:  envPrefix,
		flagSet:    flagSet,
		values:     make(map[string]interface{}),
		alreadySet: make(map[string]struct{}),
	}
}

func (c *config) RegisterString(name, description string, defaultValue string) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.String(name, defaultValue, description))
}

func (c *config) RegisterInt(name, description string, defaultValue int) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.Int(name, defaultValue, description))
}

func (c *config) RegisterUint(name, description string, defaultValue uint) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.Uint(name, defaultValue, description))
}

func (c *config) RegisterBool(name, description string, defaultValue bool) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.Bool(name, defaultValue, description))
}

func (c *config) RegisterFloat(name, description string, defaultValue float64) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.Float64(name, defaultValue, description))
}

func (c *config) RegisterDuration(name, description string, defaultValue time.Duration) error {
	name = normalizeKey(name)
	return c.register(name, description, c.flagSet.Duration(name, defaultValue, description))
}

func (c *config) GetString(name string) (string, bool) {
	variable, exists := c.get(name)
	if !exists {
		return "", false
	}

	value, ok := variable.(*string)
	if !ok {
		return "", false
	}

	return *value, true
}

func (c *config) GetStringSlice(name string) ([]string, bool) {
	variable, exists := c.GetString(name)
	if !exists {
		return nil, false
	}
	return split(variable, ","), true
}

func (c *config) GetInt(name string) (int, bool) {
	variable, exists := c.get(name)
	if !exists {
		return 0, false
	}

	value, ok := variable.(*int)
	if !ok {
		return 0, false
	}

	return *value, true
}

func (c *config) GetUint(name string) (uint, bool) {
	variable, exists := c.get(name)
	if !exists {
		return 0, false
	}

	value, ok := variable.(*uint)
	if !ok {
		return 0, false
	}

	return *value, true
}

func (c *config) GetBool(name string) (bool, bool) {
	variable, exists := c.get(name)
	if !exists {
		return false, false
	}

	value, ok := variable.(*bool)
	if !ok {
		return false, false
	}

	return *value, true
}

func (c *config) GetFloat(name string) (float64, bool) {
	variable, exists := c.get(name)
	if !exists {
		return 0, false
	}

	value, ok := variable.(*float64)
	if !ok {
		return 0, false
	}

	return *value, true
}

func (c *config) GetDuration(name string) (time.Duration, bool) {
	variable, exists := c.get(name)
	if !exists {
		return time.Duration(0), false
	}

	value, ok := variable.(*time.Duration)
	if !ok {
		return time.Duration(0), false
	}

	return *value, true
}

// String returns stringified list of flags.
func (c *config) String() string {
	var b bytes.Buffer
	b.WriteRune('[')
	c.flagSet.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(&b, "%q=%q, ", f.Name, f.Value.String())
	})
	// trim suffix ", "
	if b.Len() > 0 {
		b.Truncate(b.Len() - 2)
	}
	b.WriteRune(']')
	return b.String()
}

func (c *config) JSON() []byte {
	flagsInfo := make([]map[string]interface{}, 0)

	c.flagSet.VisitAll(func(f *flag.Flag) {
		flagsInfo = append(flagsInfo, map[string]interface{}{
			"name":          f.Name,
			"default_value": f.DefValue,
			"value":         f.Value,
			"description":   f.Usage,
		})
	})

	jsonData, _ := json.Marshal(flagsInfo)

	return jsonData
}

func (c *config) register(name, description string, value interface{}) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if _, exists := c.values[name]; exists {
		return fmt.Errorf("Option with name '%s' has already registered", name)
	}

	c.values[name] = value
	return nil
}

func (c *config) get(name string) (interface{}, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	key := normalizeKey(name)
	value, exists := c.values[key]
	return value, exists
}

// Parse parses flags.
func (c *config) Parse() (err error) {
	if !c.flagSet.Parsed() {
		err = c.flagSet.Parse(normalizeKeys(os.Args[1:]))
		if err == nil {
			c.flagSet.Visit(func(f *flag.Flag) {
				c.alreadySet[f.Name] = struct{}{}
			})
		}
	}
	return
}

func normalizeKey(key string) string {
	key = strings.Replace(key, ".", "-", -1)
	key = strings.Replace(key, "_", "-", -1)
	key = strings.ToLower(key)
	return key
}

func normalizeKeys(keys []string) (normKeys []string) {
	for _, k := range keys {
		hasValue := false
		for i := 1; i < len(k); i++ { // equals cannot be first
			if k[i] == '=' {
				hasValue = true
				normKeys = append(normKeys, normalizeKey(k[0:i])+"="+k[i+1:])
				break
			}
		}
		if !hasValue {
			normKeys = append(normKeys, normalizeKey(k))
		}
	}
	return
}

func split(configString, separator string) []string {
	if configString == `` {
		return nil
	}

	var result []string
	for _, val := range strings.Split(configString, separator) {
		val = strings.TrimSpace(val)
		if val != "" {
			result = append(result, val)
		}
	}
	return result
}
