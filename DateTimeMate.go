package DateTimeMate

import (
	"github.com/golang-module/carbon/v2"
	"strings"
)

const (
	ModName    string = "DateTimeMate"
	ModVersion string = "1.0.0"
	ModUrl     string = "https://github.com/jftuga/DateTimeMate"
)

// convertRelativeDateToActual converts "yesterday", "today", "tomorrow"
// into actual dates; yesterday and tomorrow are -/+ 24 hours of current time
func convertRelativeDateToActual(from string) string {
	switch strings.ToLower(from) {
	case "now":
		return carbon.Now().String()
	case "today":
		return carbon.Now().String()
	case "yesterday":
		return carbon.Yesterday().String()
	case "tomorrow":
		return carbon.Tomorrow().String()
	}
	return from
}
