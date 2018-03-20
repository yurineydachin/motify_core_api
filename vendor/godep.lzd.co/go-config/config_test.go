package config

import (
	"flag"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	c := newTestConfig()

	stringValue := `stringValue`
	stringName := `string`

	c.RegisterString(stringName, "test", stringValue)

	got, exists := c.GetString(stringName)
	if exists {
		if got != stringValue {
			t.Errorf("GetString returns '%s'", got)
		}
	} else {
		t.Errorf("GetString returns name '%s' doesn't exist", stringName)
	}
}

func TestConfigParseINI(t *testing.T) {
	conf := newTestConfig()

	conf.RegisterString("id", "service id", "")
	conf.RegisterUint("uint", "uint val", 0)
	conf.RegisterFloat("float", "float val", 0)
	conf.RegisterBool("logger-enabled", "logger enabled", false)
	conf.RegisterInt("server-port", "server port", 0)
	conf.RegisterDuration("cache-ttl", "cache ttl", 0)

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	id, _ := conf.GetString("id")
	if id != "test" {
		t.Fatal("id != 'test'")
	}
	u, _ := conf.GetUint("uint")
	if u != 10 {
		t.Fatal("uint != 10")
	}
	f, _ := conf.GetFloat("float")
	if f != 0.5 {
		t.Fatal("float != 0.5")
	}

	loggerEnabled, _ := conf.GetBool("logger-enabled")
	if loggerEnabled != true {
		t.Fatal("logger-enabled != true")
	}

	port, _ := conf.GetInt("server-port")
	if port != 8080 {
		t.Fatal("server-port != 8080")
	}

	ttl, _ := conf.GetDuration("cache-ttl")
	if ttl != 1*time.Minute {
		t.Fatal("cache-ttl != 1m")
	}
}

func Test_SourceHasHyphen_RegisterHyphenGetUnderscored_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("hyphen-setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	hyphen_setting, ok := conf.GetString("hyphen_setting")
	if !ok {
		t.Fatal("GetString('hyphen_setting') failed")
	}
	if hyphen_setting != "value2" {
		t.Fatal("hyphen-setting != 'value2'")
	}
}

func Test_SourceHasHyphen_RegisterHyphenGetHyphen_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("hyphen-setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	hyphen_setting, ok := conf.GetString("hyphen-setting")
	if !ok {
		t.Fatal("GetString('hyphen-setting') failed")
	}
	if hyphen_setting != "value2" {
		t.Fatal("hyphen-setting != 'value2'")
	}
}

func Test_SourceHasHyphen_RegisterUnderscoredGetHyphen_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("hyphen_setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	hyphen_setting, ok := conf.GetString("hyphen-setting")
	if !ok {
		t.Fatal("GetString('hyphen-setting') failed")
	}
	if hyphen_setting != "value2" {
		t.Fatal("hyphen_setting != 'value2'")
	}
}

func Test_SourceHasHyphen_RegisterUnderscoredGetUnderscored_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("hyphen_setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	hyphen_setting, ok := conf.GetString("hyphen_setting")
	if !ok {
		t.Fatal("GetString('hyphen_setting') failed")
	}
	if hyphen_setting != "value2" {
		t.Fatal("hyphen_setting != 'value2'")
	}
}

func Test_SourceUnderscored_RegisterUnderscoredGetHyphen_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("underscored_setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	underscored_setting, ok := conf.GetString("underscored-setting")
	if !ok {
		t.Fatal("GetString('underscored_setting') failed")
	}
	if underscored_setting != "value" {
		t.Fatal("underscored-setting != 'value'")
	}
}

func Test_SourceUnderscored_RegisterUnderscoredGetUnderscored_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("underscored_setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	underscored_setting, ok := conf.GetString("underscored_setting")
	if !ok {
		t.Fatal("GetString('underscored_setting') failed")
	}
	if underscored_setting != "value" {
		t.Fatal("underscored_setting != 'value'")
	}
}

func Test_SourceUnderscored_RegisterHyphenGetUnderscored_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("underscored-setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	underscored_setting, ok := conf.GetString("underscored_setting")
	if !ok {
		t.Fatal("GetString('underscored_setting') failed")
	}
	if underscored_setting != "value" {
		t.Fatal("underscored_setting != 'value'")
	}
}

func Test_SourceUnderscored_RegisterHyphenGetHyphen_SuccessfullyGotValue(t *testing.T) {
	conf := newTestConfig()
	conf.RegisterString("underscored-setting", "do something", "")

	if err := conf.LoadFile("test.ini"); err != nil {
		t.Fatal(err)
	}

	underscored_setting, ok := conf.GetString("underscored-setting")
	if !ok {
		t.Fatal("GetString('underscored-setting') failed")
	}
	if underscored_setting != "value" {
		t.Fatal("underscored-setting != 'value'")
	}
}

func newTestConfig() *config {
	return newConfig(flag.NewFlagSet("test", flag.PanicOnError), "test")
}
