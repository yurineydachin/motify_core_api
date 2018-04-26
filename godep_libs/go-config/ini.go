package config

import (
	"flag"
	"os"
	"strings"

	"github.com/vaughan0/go-ini"
)

// LoadFile merges ini file flags.
func (conf *config) LoadFile(filename string) error {
	if err := conf.Parse(); err != nil {
		return err
	}

	dict, err := ini.LoadFile(filename)
	if err != nil {
		return err
	}

	conf.merge(dict)
	return nil
}

func (conf *config) merge(f ini.File) {
	dict := flattenIniFile(f)

	conf.flagSet.VisitAll(func(f *flag.Flag) {
		if _, ok := conf.alreadySet[f.Name]; ok {
			return
		}

		if conf.envPrefix != "" {
			val, found := os.LookupEnv(envKey(conf.envPrefix, f.Name))
			if found {
				conf.flagSet.Set(f.Name, val)
				return
			}
		}

		val, found := dict[normalizeKey(f.Name)]
		if found {
			conf.flagSet.Set(f.Name, val)
		}
	})
}

func flattenIniFile(f ini.File) map[string]string {
	dict := make(map[string]string)
	for sectionName, section := range f {
		for k, v := range section {
			var key string
			if sectionName != "" {
				key = sectionName + "-" + k
			} else {
				key = k
			}
			key = normalizeKey(key)
			dict[key] = v
		}
	}
	return dict
}

func envKey(prefix, name string) string {
	key := prefix + "_" + name
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	return strings.ToUpper(key)
}
