package config

import (
	"flag"
	"os"
	"path"
	"time"
)

var globalConf = newConfig(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[0])

var deprecatedConfDir string

func init() {
	// TODO: remove it
	globalConf.flagSet.StringVar(&deprecatedConfDir, "config_dir", "", "DEPRECATED. It's used for backwards compatibility. Please use 'config-dir'.")

	RegisterString("config", "Single config file. Overwrites core config file 'app.ini' and env config file", "")
	RegisterString("config-dir", "directory to config files to search for 'app.ini' and env config file", "./etc/")
	RegisterString("config-env", "env config file to apply, e.g. dev for dev.ini", "")
}

func RegisterString(name, description string, defaultValue string) error {
	return globalConf.RegisterString(name, description, defaultValue)
}

func RegisterInt(name, description string, defaultValue int) error {
	return globalConf.RegisterInt(name, description, defaultValue)
}

func RegisterUint(name, description string, defaultValue uint) error {
	return globalConf.RegisterUint(name, description, defaultValue)
}

func RegisterBool(name, description string, defaultValue bool) error {
	return globalConf.RegisterBool(name, description, defaultValue)
}

func RegisterFloat(name, description string, defaultValue float64) error {
	return globalConf.RegisterFloat(name, description, defaultValue)
}

func RegisterDuration(name, description string, defaultValue time.Duration) error {
	return globalConf.RegisterDuration(name, description, defaultValue)
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

// LoadFile merges ini file flags into global config.
func LoadFile(filename string) error {
	return globalConf.LoadFile(filename)
}

// String returns stringified list of flags.
func String() string {
	return globalConf.String()
}

// JSON returns desctipions and values of flags.
func JSON() []byte {
	return globalConf.JSON()
}

// Parse parses flags.
func Parse() error {
	return globalConf.Parse()
}

// ParseAll parses flags and provided config files.
func ParseAll() error {
	if err := Parse(); err != nil {
		return err
	}

	configFile, _ := GetString("config")
	if configFile != "" {
		return loadConfigFromFile(configFile)
	}

	confDir := deprecatedConfDir
	if confDir == "" {
		confDir, _ = GetString("config-dir")
	}
	if err := loadConfigFromFile(path.Join(confDir, "app.ini")); err != nil {
		return err
	}

	// env specified config is optional
	if confEnv, ok := GetString("config-env"); ok {
		filename := path.Join(confDir, confEnv+".ini")
		if _, err := os.Stat(filename); err == nil {
			if err := loadConfigFromFile(filename); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadConfigFromFile(filename string) error {
	return globalConf.LoadFile(filename)
}
