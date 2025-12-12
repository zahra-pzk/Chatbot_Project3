package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
)

type JalaliTime time.Time

func ParseJalaliToTime(jalali string) (time.Time, error) {
	var y, m, d int
	_, err := fmt.Sscanf(jalali, "%d-%d-%d", &y, &m, &d)
	if err != nil {
		return time.Time{}, errors.New("invalid jalali date format (use YYYY-MM-DD)")
	}

	pm := ptime.Month(m)

	jt := ptime.Date(y, pm, d, 0, 0, 0, 0, time.Local)
	return jt.Time(), nil
}

func ToJalaliFormat(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	pt := ptime.New(t)
	return fmt.Sprintf("%04d-%02d-%02d", pt.Year(), pt.Month(), pt.Day())
}



func (jt JalaliTime) MarshalJSON() ([]byte, error) {
	t := time.Time(jt)
	if t.IsZero() {
		return []byte(`""`), nil
	}

	j := ptime.New(t)
	s := fmt.Sprintf("%04d-%02d-%02d", j.Year(), j.Month(), j.Day())
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

func (jt *JalaliTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		*jt = JalaliTime(time.Time{})
		return nil
	}

	var y, m, d int
	_, err := fmt.Sscanf(s, "%d-%d-%d", &y, &m, &d)
	if err != nil {
		return fmt.Errorf("invalid jalali date format: use YYYY-MM-DD")
	}

	pm := ptime.Month(m)
	t := ptime.Date(y, pm, d, 0, 0, 0, 0, time.Local).Time()
	*jt = JalaliTime(t)
	return nil
}

func (jt JalaliTime) ToTime() time.Time {
	return time.Time(jt)
}
