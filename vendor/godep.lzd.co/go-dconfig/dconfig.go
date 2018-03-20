package dconfig

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} {
	return bool(*b)
}

func (b *boolValue) String() string {
	return strconv.FormatBool(bool(*b))
}

type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} {
	return int(*i)
}

func (i *intValue) String() string {
	return strconv.Itoa(int(*i))
}

type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} {
	return int64(*i)
}

func (i *int64Value) String() string {
	return strconv.FormatInt(int64(*i), 0)
}

type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} {
	return uint(*i)
}

func (i *uintValue) String() string {
	return strconv.FormatUint(uint64(uint(*i)), 10)
}

type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} {
	return string(*s)
}

func (s *stringValue) String() string {
	return string(*s)
}

type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string {
	return strconv.FormatFloat(float64(*f), 'g', -1, 64)
}

type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

type config struct {
	values map[string]*Var
	mtx    sync.RWMutex
}

type Var struct {
	Name        string // name as it appears on command line
	Description string // help message
	Value       Value  // value as set
	Action      func(val interface{})
}

type Value interface {
	String() string
	Set(string) error
	Get() interface{}
}

func newConfig() *config {
	return &config{
		values: make(map[string]*Var),
	}
}

func (conf *config) RegisterBool(name string, value bool, description string, action func(val bool)) {
	var b bool
	conf.register(newBoolValue(value, &b), name, description, func(val interface{}) {
		action(val.(bool))
	})
}

func (conf *config) RegisterInt(name string, value int, description string, action func(val int)) {
	var i int
	conf.register(newIntValue(value, &i), name, description, func(val interface{}) {
		action(val.(int))
	})
}

func (conf *config) Int64(name string, value int64, description string, action func(val int64)) {
	var i int64
	conf.register(newInt64Value(value, &i), name, description, func(val interface{}) {
		action(val.(int64))
	})
}

func (conf *config) RegisterUint(name string, value uint, description string, action func(val uint)) {
	var i uint
	conf.register(newUintValue(value, &i), name, description, func(val interface{}) {
		action(val.(uint))
	})
}

func (conf *config) Uint64(name string, value uint64, description string, action func(val uint64)) {
	var i uint64
	conf.register(newUint64Value(value, &i), name, description, func(val interface{}) {
		action(val.(uint64))
	})
}

func (conf *config) RegisterString(name string, value string, description string, action func(val string)) {
	var s string
	conf.register(newStringValue(value, &s), name, description, func(val interface{}) {
		action(val.(string))
	})
}

func (conf *config) RegisterFloat(name string, value float64, description string, action func(val float64)) {
	var f float64
	conf.register(newFloat64Value(value, &f), name, description, func(val interface{}) {
		action(val.(float64))
	})
}

func (conf *config) RegisterDuration(name string, value time.Duration, description string, action func(val time.Duration)) {
	var d time.Duration
	conf.register(newDurationValue(value, &d), name, description, func(val interface{}) {
		action(val.(time.Duration))
	})
}

func (conf *config) GetString(name string) (string, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return "", false
	}
	return variable.Value.String(), true
}

func (conf *config) GetStringSlice(name string) ([]string, bool) {
	variable, exists := conf.GetString(name)
	if !exists {
		return nil, false
	}
	return split(variable, ","), true
}

func (conf *config) GetInt(name string) (int, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return 0, false
	}
	return variable.Value.Get().(int), true
}

func (conf *config) GetUint(name string) (uint, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return 0, false
	}
	return variable.Value.Get().(uint), true
}

func (conf *config) GetBool(name string) (bool, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return false, false
	}
	return variable.Value.Get().(bool), true
}

func (conf *config) GetFloat(name string) (float64, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return 0, false
	}
	return variable.Value.Get().(float64), true
}

func (conf *config) GetDuration(name string) (time.Duration, bool) {
	variable, exists := conf.get(name)
	if !exists {
		return time.Duration(0), false
	}
	return variable.Value.Get().(time.Duration), true
}

func (conf *config) register(value Value, name string, description string, action func(val interface{})) {
	conf.mtx.Lock()
	defer conf.mtx.Unlock()

	if _, alreadythere := conf.values[name]; alreadythere {
		panic(fmt.Sprintf("var redefined: %s", name))
	}
	conf.values[name] = &Var{
		Name:        name,
		Value:       value,
		Description: description,
		Action:      action,
	}
}

func (conf *config) get(name string) (*Var, bool) {
	conf.mtx.RLock()
	defer conf.mtx.RUnlock()

	v, ok := conf.values[name]
	return v, ok
}

func (conf *config) set(name, value string) error {
	v, changed, err := conf.syncSet(name, value)
	if err != nil {
		return err
	}
	if changed {
		v.Action(v.Value.Get())
	}
	return nil
}

func (conf *config) syncSet(name, value string) (*Var, bool, error) {
	conf.mtx.Lock()
	defer conf.mtx.Unlock()

	v, ok := conf.values[name]
	if !ok {
		return nil, false, fmt.Errorf("no such var %q", name)
	}
	if v.Value.String() == value {
		return v, false, nil
	}
	if err := v.Value.Set(value); err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func (conf *config) visit(fn func(name string, v *Var)) {
	conf.mtx.RLock()
	defer conf.mtx.RUnlock()

	for name, v := range conf.values {
		fn(name, v)
	}
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
