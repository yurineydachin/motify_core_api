package main

import (
	"time"

	"godep.lzd.co/go-config"
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
