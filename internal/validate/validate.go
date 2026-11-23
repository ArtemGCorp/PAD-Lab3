// Package validate checks user input.
package validate

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MaxAmount is the largest accepted amount.
const MaxAmount = 1e12

var (
	ErrEmpty    = errors.New("введите сумму")
	ErrNotNum   = errors.New("сумма должна быть числом, например 100 или 12,50")
	ErrNotPos   = errors.New("сумма должна быть больше нуля")
	ErrTooLarge = errors.New("слишком большая сумма")

	ErrDateEmpty   = errors.New("введите дату в формате ДД.ММ.ГГГГ")
	ErrDateFormat  = errors.New("дата должна быть в формате ДД.ММ.ГГГГ, например 05.03.2026")
	ErrDateInvalid = errors.New("такой даты не существует")
)

var amountRe = regexp.MustCompile(`^\d+([.,]\d+)?$`)

// ParseAmount accepts "100", "12.5", "12,50" (spaces around are trimmed,
// spaces inside are thousand separators: "1 000"). Rejects empty, letters,
// signs, exponents, zero, values above MaxAmount.
func ParseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrEmpty
	}
	s = strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), " ", "")
	if strings.HasPrefix(s, "-") {
		if _, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64); err == nil {
			return 0, ErrNotPos
		}
	}
	if !amountRe.MatchString(s) {
		return 0, ErrNotNum
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, ErrNotNum
	}
	if v <= 0 {
		return 0, ErrNotPos
	}
	if v > MaxAmount {
		return 0, ErrTooLarge
	}
	return v, nil
}

var dateRe = regexp.MustCompile(`^(\d{2})\.(\d{2})\.(\d{4})$`)

// ParseDate parses "DD.MM.YYYY" and rejects non-existent days (31.02).
func ParseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, ErrDateEmpty
	}
	m := dateRe.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, ErrDateFormat
	}
	d, _ := strconv.Atoi(m[1])
	mo, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])
	t := time.Date(y, time.Month(mo), d, 0, 0, 0, 0, time.Local)
	if t.Day() != d || int(t.Month()) != mo || t.Year() != y {
		return time.Time{}, ErrDateInvalid
	}
	return t, nil
}
