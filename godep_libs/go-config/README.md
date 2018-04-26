# motify_core_api/godep_libs/go-config
## Description
The library combines \*.ini files and flags. You work with config parameters like with go flags.
It supports groups in \*.ini files.

## Example
**main.go**
```go
package main

import (
	"time"

	"motify_core_api/godep_libs/go-config"
)

func init() {
	if err := config.RegisterString("str-value", "String value", "test"); err != nil {
		panic(err)
	}
	if err := config.RegisterDuration("duration-value", "Duratuib value", time.Second*5); err != nil {
		panic(err)
	}

	if err := config.RegisterInt("group1-int", "Int value in group", 123); err != nil {
		panic(err)
	}

	if err := config.RegisterBool("group1-bool", "Bool value in group", false); err != nil {
		panic(err)
	}
}

func main() {
	if err := config.ParseAll(); err != nil {
		panic(err)
	}

	strVal, _ := config.GetString("str-value")
	durVal, _ := config.GetDuration("duration-value")
	groupInt, _ := config.GetInt("group1-int")
	groupBool, _ := config.GetBool("group1-bool")

	println(strVal)
	println(durVal.String())
	println(groupInt)
	println(groupBool)
}
```
**app.ini**
```ini
str-value       = value in config

; You can use "_" instead of "-: in config file
duration_value  = 10s

[group1]
int             = 456
bool            = true
```

## How it works
```
$ go run main.go -help
Usage of /tmp/go-build900796515/command-line-arguments/_obj/exe/main:
  -config string
    	Single config file. Overwrites core config file 'app.ini' and env config file
  -config-dir string
    	directory to config files to search for 'app.ini' and env config file (default "./")
  -config-env string
    	env config file to apply, e.g. dev for dev.ini
  -config_dir string
    	DEPRECATED. It's used for backwards compatibility. Please use 'config-dir'.
  -duration-value duration
    	Duratuib value (default 5s)
  -group1-bool
    	Bool value in group
  -group1-int int
    	Int value in group (default 123)
  -str-value string
    	String value (default "test")
```

```
$ go run main.go
value in config
10s
456
true
```