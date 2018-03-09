package config

import (
//"fmt"
	"flag"
	"os"
	"strings"

//	"godep.lzd.co/service/logger"
	"github.com/vaughan0/go-ini"
)

// LoadFile merges ini file flags.
func (conf *config) LoadFile(filename string) error {
    //logger.Error(nil, "filename %s", filename)
	if err := conf.Parse(); err != nil {
		return err
	}

	dict, err := ini.LoadFile(filename)
    //logger.Error(nil, "loadFile %s, data %#v", filename, dict)
	if err != nil {
		return err
	}

	conf.merge(dict)
	return nil
}

func (conf *config) merge(f ini.File) {
    //logger.Error(nil, "config merge f ini.File: %#v", f)
	dict := flattenIniFile(f)
    //logger.Error(nil, "config merge flattenIniFile(f): %#v", dict)

	conf.flagSet.VisitAll(func(f *flag.Flag) {
		if _, ok := conf.alreadySet[f.Name]; ok {
        //logger.Error(nil, "config merge alreadySet: %s", f.Name)
			return
		}

		if conf.envPrefix != "" {
			val, found := os.LookupEnv(envKey(conf.envPrefix, f.Name))
			if found {
        //logger.Error(nil, "envKey conf.flagSet.Set(%s, %#v)", f.Name, val)
				conf.flagSet.Set(f.Name, val)
				return
			} else {
        //logger.Error(nil, "not found envKey conf.flagSet.Set(%s, %#v)", f.Name, val)
            }
		}

		val, found := dict[normalizeKey(f.Name)]
        //logger.Error(nil, "try dict[normalizeKey(f.Name) : %s] ", normalizeKey(f.Name))
        //logger.Error(nil, "try dict[normalizeKey(f.Name)] : %#v ", found)
		if found {
        //logger.Error(nil, fmt.Sprintf("normalizeKey conf.flagSet.Set(%s, %#v), normalizeKey(f.Name): %s  ", f.Name, val, normalizeKey(f.Name)))
			conf.flagSet.Set(f.Name, val)
		} else {
        //logger.Error(nil, fmt.Sprintf("not found normalizeKey conf.flagSet.Set(%s, %#v), normalizeKey(f.Name): %s  ", f.Name, val, normalizeKey(f.Name)))
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
