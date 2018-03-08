package common

import "strings"

const (
	listSeparator = ","
)

// StringList is a type for string values, separated by listSeparator
type StringList []string

// String converts values to string
func (l *StringList) String() string {
	return strings.Join(*l, listSeparator)
}

// Set converts values from string
func (l *StringList) Set(value string) error {
	newValues := StringList{}
	for _, item := range strings.Split(value, listSeparator) {
		item = strings.TrimSpace(item)
		if len(item) > 0 {
			newValues = append(newValues, item)
		}
	}
	*l = newValues
	return nil
}
