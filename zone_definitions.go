package DateTimeMate

// LoadZoneDefinitions returns a map of timezone abbreviations to their UTC offsets
func LoadZoneDefinitions() map[string]int {
	return map[string]int{
		// North America
		"PST":  -28800, // Pacific Standard Time (UTC-8)
		"PDT":  -25200, // Pacific Daylight Time (UTC-7)
		"MST":  -25200, // Mountain Standard Time (UTC-7)
		"MDT":  -21600, // Mountain Daylight Time (UTC-6)
		"CST":  -21600, // Central Standard Time (UTC-6) [North America]
		"CDT":  -18000, // Central Daylight Time (UTC-5)
		"EST":  -18000, // Eastern Standard Time (UTC-5)
		"EDT":  -14400, // Eastern Daylight Time (UTC-4)
		"AST":  -14400, // Atlantic Standard Time (UTC-4)
		"ADT":  -10800, // Atlantic Daylight Time (UTC-3)
		"NST":  -12600, // Newfoundland Standard Time (UTC-3:30)
		"NDT":  -9000,  // Newfoundland Daylight Time (UTC-2:30)
		"HAT":  -9000,  // Newfoundland Daylight Time (Same as NDT)
		"AKST": -32400, // Alaska Standard Time (UTC-9)
		"AKDT": -28800, // Alaska Daylight Time (UTC-8)
		"HST":  -36000, // Hawaii Standard Time (UTC-10)
		"HAST": -36000, // Hawaii-Aleutian Standard Time (Same as HST)
		"HADT": -32400, // Hawaii-Aleutian Daylight Time (UTC-9)

		// Europe
		"WET":  0,     // Western European Time (UTC+0)
		"WEST": 3600,  // Western European Summer Time (UTC+1)
		"CET":  3600,  // Central European Time (UTC+1)
		"CEST": 7200,  // Central European Summer Time (UTC+2)
		"EET":  7200,  // Eastern European Time (UTC+2)
		"EEST": 10800, // Eastern European Summer Time (UTC+3)
		"BST":  3600,  // British Summer Time (UTC+1)
		"IRST": 3600,  // Irish Standard Time (UTC+1)

		// Asia & Oceania
		"IST":  19800, // India Standard Time (UTC+5:30)
		"HKT":  28800, // Hong Kong Time (UTC+8)
		"SGT":  28800, // Singapore Time (UTC+8)
		"JST":  32400, // Japan Standard Time (UTC+9)
		"KST":  32400, // Korea Standard Time (UTC+9)
		"PHT":  28800, // Philippine Time (UTC+8)
		"CNST": 28800, // China Standard Time (UTC+8)
		"SST":  28800, // Singapore Standard Time (UTC+8)

		// Australia
		"AWST": 28800, // Australian Western Standard Time (UTC+8)
		"ACST": 34200, // Australian Central Standard Time (UTC+9:30)
		"AEST": 36000, // Australian Eastern Standard Time (UTC+10)
		"ACDT": 37800, // Australian Central Daylight Time (UTC+10:30)
		"AEDT": 39600, // Australian Eastern Daylight Time (UTC+11)

		// New Zealand
		"NZST": 43200, // New Zealand Standard Time (UTC+12)
		"NZDT": 46800, // New Zealand Daylight Time (UTC+13)

		// South America
		"ART":  -10800, // Argentina Time (UTC-3)
		"BRT":  -10800, // Brasilia Time (UTC-3)
		"BRST": -7200,  // Brasilia Summer Time (UTC-2)

		// Middle East
		"EAT": 10800, // East Africa Time (UTC+3)
		"GST": 14400, // Gulf Standard Time (UTC+4)
		"PKT": 18000, // Pakistan Standard Time (UTC+5)

		// Military/NATO time zones
		"Z": 0,      // Zulu Time (UTC)
		"A": 3600,   // Alpha Time Zone (UTC+1)
		"B": 7200,   // Bravo Time Zone (UTC+2)
		"C": 10800,  // Charlie Time Zone (UTC+3)
		"D": 14400,  // Delta Time Zone (UTC+4)
		"E": 18000,  // Echo Time Zone (UTC+5)
		"F": 21600,  // Foxtrot Time Zone (UTC+6)
		"G": 25200,  // Golf Time Zone (UTC+7)
		"H": 28800,  // Hotel Time Zone (UTC+8)
		"I": 32400,  // India Time Zone (UTC+9)
		"K": 36000,  // Kilo Time Zone (UTC+10)
		"L": 39600,  // Lima Time Zone (UTC+11)
		"M": 43200,  // Mike Time Zone (UTC+12)
		"N": -3600,  // November Time Zone (UTC-1)
		"O": -7200,  // Oscar Time Zone (UTC-2)
		"P": -10800, // Papa Time Zone (UTC-3)
		"Q": -14400, // Quebec Time Zone (UTC-4)
		"R": -18000, // Romeo Time Zone (UTC-5)
		"S": -21600, // Sierra Time Zone (UTC-6)
		"T": -25200, // Tango Time Zone (UTC-7)
		"U": -28800, // Uniform Time Zone (UTC-8)
		"V": -32400, // Victor Time Zone (UTC-9)
		"W": -36000, // Whiskey Time Zone (UTC-10)
		"X": -39600, // X-ray Time Zone (UTC-11)
		"Y": -43200, // Yankee Time Zone (UTC-12)

		// Generic
		"UTC": 0, // Coordinated Universal Time
		"GMT": 0, // Greenwich Mean Time
		"UT":  0, // Universal Time
	}
}
