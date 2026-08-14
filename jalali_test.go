package main

import (
	"testing"
	"time"
)

func TestJalaliDate(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"regular date", time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC), "۱۴۰۵/۰۵/۲۳"},
		{"nowruz 1403", time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), "۱۴۰۳/۰۱/۰۱"},
		{"nowruz 1404", time.Date(2025, 3, 21, 0, 0, 0, 0, time.UTC), "۱۴۰۴/۰۱/۰۱"},
		{"unix epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), "۱۳۴۸/۱۰/۱۱"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jalaliDate(tt.date); got != tt.want {
				t.Errorf("jalaliDate(%v) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

