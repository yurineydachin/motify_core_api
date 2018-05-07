package dconfig

import "time"

var globalConf = newConfig()

func RegisterString(name, description string, defaultValue string, action func(val string)) {
	globalConf.RegisterString(name, defaultValue, description, action)
}

func RegisterInt(name, description string, defaultValue int, action func(val int)) {
	globalConf.RegisterInt(name, defaultValue, description, action)
}

func RegisterUint(name, description string, defaultValue uint, action func(val uint)) {
	globalConf.RegisterUint(name, defaultValue, description, action)
}

func RegisterBool(name, description string, defaultValue bool, action func(val bool)) {
	globalConf.RegisterBool(name, defaultValue, description, action)
}

func RegisterFloat(name, description string, defaultValue float64, action func(val float64)) {
	globalConf.RegisterFloat(name, defaultValue, description, action)
}

func RegisterDuration(name, description string, defaultValue time.Duration, action func(val time.Duration)) {
	globalConf.RegisterDuration(name, defaultValue, description, action)
}

func GetString(name string) (string, bool) {
	return globalConf.GetString(name)
}

func GetStringSlice(name string) ([]string, bool) {
	return globalConf.GetStringSlice(name)
}

func GetInt(name string) (int, bool) {
	return globalConf.GetInt(name)
}

func GetUint(name string) (uint, bool) {
	return globalConf.GetUint(name)
}

func GetBool(name string) (bool, bool) {
	return globalConf.GetBool(name)
}

func GetFloat(name string) (float64, bool) {
	return globalConf.GetFloat(name)
}

func GetDuration(name string) (time.Duration, bool) {
	return globalConf.GetDuration(name)
}
