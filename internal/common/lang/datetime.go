package lang

import (
	"strings"
	"time"
)

// Java's java.time.LocalDateTime carries no zone. Spring Boot 2.x registers
// JavaTimeModule and disables WRITE_DATES_AS_TIMESTAMPS, so a bare
// LocalDateTime serialises as ISO-8601 without an offset, and Jackson omits a
// zero seconds field only when nanoseconds are also zero - it always writes
// seconds when they are non-zero. The MySQL DATETIME columns in vueblog.sql
// have second precision, so the emitted form is yyyy-MM-ddTHH:mm:ss.
const localDateTimeLayout = "2006-01-02T15:04:05"

// LocalDateTimeLayout is exported for tests and the MySQL driver.
const LocalDateTimeLayout = localDateTimeLayout

// localDateLayout matches @JsonFormat(pattern="yyyy-MM-dd") on Blog.created.
const localDateLayout = "2006-01-02"

// LocalDateTime mirrors an unannotated java.time.LocalDateTime field.
// A nil pointer marshals to null, matching a null Java field under Jackson's
// ALWAYS inclusion.
type LocalDateTime struct {
	T time.Time
}

func NewLocalDateTime(t time.Time) *LocalDateTime { return &LocalDateTime{T: t} }

func (l LocalDateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.T.Format(localDateTimeLayout) + `"`), nil
}

func (l *LocalDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.ParseInLocation(localDateTimeLayout, s, time.Local)
	if err != nil {
		return err
	}
	l.T = t
	return nil
}

func (l LocalDateTime) String() string { return l.T.Format(localDateTimeLayout) }

// LocalDate mirrors a java.time.LocalDateTime field annotated with
// @JsonFormat(pattern="yyyy-MM-dd"): the value keeps full precision in memory
// (Blog.created is compared and stored as a datetime) but serialises date-only.
type LocalDate struct {
	T time.Time
}

func NewLocalDate(t time.Time) *LocalDate { return &LocalDate{T: t} }

func (l LocalDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + l.T.Format(localDateLayout) + `"`), nil
}

// UnmarshalJSON accepts both the date-only form Jackson emits and the full
// datetime form, because @JsonFormat applies to deserialisation too but the
// database round-trip carries the time component.
func (l *LocalDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}
	if t, err := time.ParseInLocation(localDateTimeLayout, s, time.Local); err == nil {
		l.T = t
		return nil
	}
	t, err := time.ParseInLocation(localDateLayout, s, time.Local)
	if err != nil {
		return err
	}
	l.T = t
	return nil
}

func (l LocalDate) String() string { return l.T.Format(localDateTimeLayout) }
