package DateTimeMate

// ZoneDefinition describes a time zone abbreviation: its fixed UTC offset
// in seconds, a human readable description, and, for abbreviations that
// carry more than one real-world meaning, the other interpretations
type ZoneDefinition struct {
	Offset      int
	Description string
	Ambiguous   string
}

// LoadZoneDefinitions returns a map of timezone abbreviations to their definitions
func LoadZoneDefinitions() map[string]ZoneDefinition {
	return map[string]ZoneDefinition{
		// North America
		"PST":  {-28800, "Pacific Standard Time", "Philippine Standard Time (UTC+8)"},
		"PDT":  {-25200, "Pacific Daylight Time", ""},
		"MST":  {-25200, "Mountain Standard Time", ""},
		"MDT":  {-21600, "Mountain Daylight Time", ""},
		"CST":  {-21600, "Central Standard Time (North America)", "China Standard Time (UTC+8), Cuba Standard Time (UTC-5)"},
		"CDT":  {-18000, "Central Daylight Time (North America)", "Cuba Daylight Time (UTC-4)"},
		"EST":  {-18000, "Eastern Standard Time (North America)", "Australian Eastern Standard Time (UTC+10)"},
		"EDT":  {-14400, "Eastern Daylight Time", ""},
		"AST":  {-14400, "Atlantic Standard Time", "Arabia Standard Time (UTC+3)"},
		"ADT":  {-10800, "Atlantic Daylight Time", ""},
		"NST":  {-12600, "Newfoundland Standard Time", ""},
		"NDT":  {-9000, "Newfoundland Daylight Time", ""},
		"HAT":  {-9000, "Newfoundland Daylight Time (same as NDT)", ""},
		"AKST": {-32400, "Alaska Standard Time", ""},
		"AKDT": {-28800, "Alaska Daylight Time", ""},
		"HST":  {-36000, "Hawaii Standard Time", ""},
		"HAST": {-36000, "Hawaii-Aleutian Standard Time (same as HST)", ""},
		"HADT": {-32400, "Hawaii-Aleutian Daylight Time", ""},

		// Europe
		"WET":  {0, "Western European Time", ""},
		"WEST": {3600, "Western European Summer Time", ""},
		"CET":  {3600, "Central European Time", ""},
		"CEST": {7200, "Central European Summer Time", ""},
		"EET":  {7200, "Eastern European Time", ""},
		"EEST": {10800, "Eastern European Summer Time", ""},
		"BST":  {3600, "British Summer Time", "Bangladesh Standard Time (UTC+6)"},

		// Asia
		"IST":  {19800, "India Standard Time", "Israel Standard Time (UTC+2), Irish Standard Time (UTC+1)"},
		"HKT":  {28800, "Hong Kong Time", ""},
		"SGT":  {28800, "Singapore Time", ""},
		"JST":  {32400, "Japan Standard Time", ""},
		"KST":  {32400, "Korea Standard Time", ""},
		"PHT":  {28800, "Philippine Time", ""},
		"CNST": {28800, "China Standard Time", ""},
		"NPT":  {20700, "Nepal Time", ""},

		// Australia
		"AWST":  {28800, "Australian Western Standard Time", ""},
		"ACST":  {34200, "Australian Central Standard Time", ""},
		"AEST":  {36000, "Australian Eastern Standard Time", ""},
		"ACDT":  {37800, "Australian Central Daylight Time", ""},
		"AEDT":  {39600, "Australian Eastern Daylight Time", ""},
		"ACWST": {31500, "Australian Central Western Standard Time", ""},
		"LHDT":  {39600, "Lord Howe Daylight Time (unusual 30-minute DST advance)", ""},
		"CAST":  {34200, "Central Australia Standard Time (synonym for ACST)", ""},

		// New Zealand & Pacific
		"NZST":  {43200, "New Zealand Standard Time", ""},
		"NZDT":  {46800, "New Zealand Daylight Time", ""},
		"CHAST": {45900, "Chatham Standard Time", ""},
		"SST":   {-39600, "Samoa Standard Time", "Singapore Standard Time (UTC+8; use SGT)"},

		// South America
		"ART":  {-10800, "Argentina Time", ""},
		"BRT":  {-10800, "Brasilia Time", ""},
		"BRST": {-7200, "Brasilia Summer Time", ""},

		// Africa & Middle East
		"EAT":  {10800, "East Africa Time", ""},
		"GST":  {14400, "Gulf Standard Time", "South Georgia Time (UTC-2)"},
		"PKT":  {18000, "Pakistan Standard Time", ""},
		"IRST": {12600, "Iran Standard Time", ""},
		"AFT":  {16200, "Afghanistan Time", ""},

		// Generic
		"UTC": {0, "Coordinated Universal Time", ""},
		"GMT": {0, "Greenwich Mean Time", ""},
		"UT":  {0, "Universal Time", ""},
	}
}
