package main

import (
	"fmt"
	"strings"
	"time"
)

// jalaliDate returns the Persian (Jalali) calendar date of t as YYYY/MM/DD
// using Persian digits, e.g. ۱۴۰۵/۰۵/۲۳.
func jalaliDate(t time.Time) string {
	jy, jm, jd := toJalali(t.Year(), int(t.Month()), t.Day())
	return toPersianDigits(fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd))
}

// toJalali converts a Gregorian date to the Jalali calendar using the
// astronomical algorithm by Kazimierz M. Borkowski, valid for the range
// 1900-2100.
func toJalali(gy, gm, gd int) (jy, jm, jd int) {
	daysInGregorianMonths := [...]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}

	if gy > 1600 {
		jy = 979
		gy -= 1600
	} else {
		jy = 0
		gy -= 621
	}

	gy2 := gy
	if gm > 2 {
		gy2 = gy + 1
	}

	days := 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 - 80 +
		gd + daysInGregorianMonths[gm-1]

	jy += 33 * (days / 12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}

	if days < 186 {
		jm = 1 + days/31
		jd = 1 + days%31
	} else {
		jm = 7 + (days-186)/30
		jd = 1 + (days-186)%30
	}
	return jy, jm, jd
}

// toPersianDigits replaces ASCII digits with their Persian equivalents.
func toPersianDigits(s string) string {
	return strings.NewReplacer(
		"0", "۰", "1", "۱", "2", "۲", "3", "۳", "4", "۴",
		"5", "۵", "6", "۶", "7", "۷", "8", "۸", "9", "۹",
	).Replace(s)
}
