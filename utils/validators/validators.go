package validators

import (
	"net/url"
	"regexp"
	"time"
)

var (
	reEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	rePhone = regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
)

func IsValidEmail(value string) bool {
	return reEmail.MatchString(value)
}

func IsValidPhone(value string) bool {
	return rePhone.MatchString(value)
}

func IsValidUrl(value string) bool {
	_, err := url.ParseRequestURI(value)
	return err == nil
}

func IsValidDatetime(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}
