package utils

import (
	"time"
)

const JSONTimeFormat = `"2006-01-02 15:04:05"`

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(JSONTimeFormat)
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(b []byte) error {
	tmp, err := time.Parse(JSONTimeFormat, string(b))
	if err != nil {
		return err
	}
	*t = JSONTime(tmp)

	return nil
}

func (t JSONTime) Add(d time.Duration) JSONTime {
	tmp := time.Time(t)

	tmp = tmp.Add(d)
	return JSONTime(tmp)
}

func (t JSONTime) Unix() int64 {
	return time.Time(t).Unix()
}
