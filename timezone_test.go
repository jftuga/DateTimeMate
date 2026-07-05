package DateTimeMate

import (
	"fmt"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
)

func setupConverter(t *testing.T) *TimeZoneConverter {
	defaultZones := LoadZoneDefinitions()
	conv, err := NewTimeZoneConverter(TimeZoneConverterWithZoneAbbrevs(defaultZones))
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}
	return conv
}

func TestTimezoneBasicConversions(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "UTC to EST",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "EST",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "PST to JST",
			sourceTime: "2024-01-15 09:00:00 PST",
			targetZone: "JST",
			expected:   "2024-01-16 02:00:00 JST",
		},
		{
			name:       "EST to GMT",
			sourceTime: "2024-01-15 17:30:00 EST",
			targetZone: "GMT",
			expected:   "2024-01-15 22:30:00 GMT",
		},
	}

	for _, tt := range tests {
		println(tt.name, tt.sourceTime)
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			println("          ", result.String())
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezoneHalfHourOffsets(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "UTC to India (UTC+5:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "IST",
			expected:   "2024-01-15 17:30:00 IST",
		},
		{
			name:       "UTC to India (UTC+5:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Kolkata",
			expected:   "2024-01-15 17:30:00 Asia/Kolkata",
		},
		{
			name:       "UTC to Iran (UTC+3:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "IRST",
			expected:   "2024-01-15 15:30:00 IRST",
		},
		{
			name:       "UTC to Iran (UTC+3:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Tehran",
			expected:   "2024-01-15 15:30:00 Asia/Tehran",
		},
		{
			name:       "UTC to Newfoundland (UTC-3:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "NST",
			expected:   "2024-01-15 08:30:00 NST",
		},
		{
			name:       "UTC to Afghanistan (UTC+4:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "AFT",
			expected:   "2024-01-15 16:30:00 AFT",
		},
		{
			name:       "UTC to Afghanistan (UTC+4:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Kabul",
			expected:   "2024-01-15 16:30:00 Asia/Kabul",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezoneQuarterHourOffsets(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "UTC to Nepal (UTC+5:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "NPT",
			expected:   "2024-01-15 17:45:00 NPT",
		},
		{
			name:       "UTC to Nepal (UTC+5:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Kathmandu",
			expected:   "2024-01-15 17:45:00 Asia/Kathmandu",
		},
		{
			name:       "UTC to Chatham Islands (UTC+12:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "CHAST",
			expected:   "2024-01-16 00:45:00 CHAST",
		},
		{
			name:       "UTC to Chatham Islands (UTC+12:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Pacific/Chatham",
			expected:   "2024-01-16 00:45:00 Pacific/Chatham",
		},
		{
			name:       "UTC to Eucla Australia (UTC+8:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "ACWST",
			expected:   "2024-01-15 20:45:00 ACWST",
		},
		{
			name:       "UTC to Eucla Australia (UTC+8:45)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Australia/Eucla",
			expected:   "2024-01-15 20:45:00 Australia/Eucla",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezoneSpecialCases(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Arizona (No DST)",
			sourceTime: "2024-07-15 12:00:00 UTC", // Summer
			targetZone: "MST",
			expected:   "2024-07-15 05:00:00 MST",
		},
		{
			name:       "Arizona (No DST)",
			sourceTime: "2024-07-15 12:00:00 UTC", // Summer
			targetZone: "America/Phoenix",
			expected:   "2024-07-15 05:00:00 America/Phoenix",
		},
		{
			name:       "Indiana (Eastern Time)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "EST",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Indiana (Eastern Time)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Indianapolis",
			expected:   "2024-01-15 07:00:00 America/Indiana/Indianapolis",
		},
		{
			name:       "Lord Howe Island (UTC+10:30/+11)",
			sourceTime: "2024-01-15 12:00:00 UTC", // Summer
			targetZone: "LHDT",
			expected:   "2024-01-15 23:00:00 LHDT", // FIXME
		},
		{
			name:       "Lord Howe Island (UTC+10:30/+11)",
			sourceTime: "2024-01-15 12:00:00 UTC", // Summer
			targetZone: "Australia/Lord_Howe",
			expected:   "2024-01-15 22:30:00 Australia/Lord_Howe", // FIXME
		},
		//{
		//	name:       "Antarctica/Casey",
		//	sourceTime: "2024-01-15 12:00:00 UTC",
		//	targetZone: "CAST",
		//	expected:   "2024-01-15 19:00:00 CAST", // FIXME
		//},
		//{
		//	name:       "Antarctica/Casey",
		//	sourceTime: "2024-01-15 12:00:00 UTC",
		//	targetZone: "Antarctica/Casey",
		//	expected:   "2024-01-15 19:00:00 Antarctica/Casey", // FIXME
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

//func TestTimezoneNumericOffsets(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expected   string
//	}{
//		{
//			name:       "UTC to +0530",
//			sourceTime: "2024-01-15 12:00:00 UTC",
//			targetZone: "19800", // +5:30 in seconds
//			expected:   "2024-01-15 17:30:00 UTC+5",
//		},
//		{
//			name:       "UTC to -0930",
//			sourceTime: "2024-01-15 12:00:00 UTC",
//			targetZone: "-34200", // -9:30 in seconds
//			expected:   "2024-01-15 02:30:00 UTC-9",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			assert.NoError(t, err)
//			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
//		})
//	}
//}

// FIXME - add this function back in
//func TestTimezoneErrorCases(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expectErr  bool
//	}{
//		{
//			name:       "Invalid source time format",
//			sourceTime: "2024-13-45 99:99:99 XYZ",
//			targetZone: "UTC",
//			expectErr:  true,
//		},
//		{
//			name:       "Unknown timezone",
//			sourceTime: "2024-01-15 12:00:00 UTC",
//			targetZone: "INVALID_TZ",
//			expectErr:  true,
//		},
//		{
//			name:       "Invalid offset value",
//			sourceTime: "2024-01-15 12:00:00 UTC",
//			targetZone: "999999",
//			expectErr:  true,
//		},
//		{
//			name:       "Empty input",
//			sourceTime: "",
//			targetZone: "UTC",
//			expectErr:  true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			if tt.expectErr {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//			}
//		})
//	}
//}

// FIXME add this function back it - this is complicated
//func TestTimezoneDSTTransitions(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expected   string
//	}{
//		{
//			name:       "US DST Spring Forward",
//			sourceTime: "2024-03-10 06:00:00 UTC",
//			targetZone: "EDT",
//			expected:   "2024-03-10 02:00:00 EDT",
//		},
//		{
//			name:       "US DST Spring Forward",
//			sourceTime: "2024-03-10 06:00:00 UTC",
//			targetZone: "America/New_York",
//			expected:   "2024-03-10 02:00:00 America/New_York",
//		},
//		//{
//		//	name:       "US DST Fall Back",
//		//	sourceTime: "2024-11-03 06:00:00 UTC",
//		//	targetZone: "America/New_York",
//		//	expected:   "2024-11-03 01:00:00 EST",
//		//},
//		//{
//		//	name:       "European DST Start",
//		//	sourceTime: "2024-03-31 01:00:00 UTC",
//		//	targetZone: "Europe/Paris",
//		//	expected:   "2024-03-31 03:00:00 CEST",
//		//},
//		//{
//		//	name:       "European DST End",
//		//	sourceTime: "2024-10-27 01:00:00 UTC",
//		//	targetZone: "Europe/Paris",
//		//	expected:   "2024-10-27 02:00:00 CET",
//		//},
//		//{
//		//	name:       "Southern Hemisphere DST Start",
//		//	sourceTime: "2024-10-06 16:00:00 UTC",
//		//	targetZone: "Pacific/Auckland",
//		//	expected:   "2024-10-07 05:00:00 NZDT",
//		//},
//		//{
//		//	name:       "Southern Hemisphere DST End",
//		//	sourceTime: "2024-04-07 14:00:00 UTC",
//		//	targetZone: "Pacific/Auckland",
//		//	expected:   "2024-04-08 02:00:00 NZST",
//		//},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			assert.NoError(t, err)
//			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
//		})
//	}
//}

//func TestTimezoneHistoricalTimezones(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expected   string
//	}{
//		{
//			name:       "Historical Shanghai Time",
//			sourceTime: "1927-01-01 12:00:00 UTC",
//			targetZone: "Asia/Shanghai",
//			expected:   "1927-01-01 19:45:00 CST",
//		},
//		{
//			name:       "Pre-2000 Israel",
//			sourceTime: "1995-01-01 12:00:00 UTC",
//			targetZone: "Asia/Jerusalem",
//			expected:   "1995-01-01 14:00:00 IST",
//		},
//		{
//			name:       "Pre-1986 Eastern Australia",
//			sourceTime: "1985-01-01 12:00:00 UTC",
//			targetZone: "Australia/Sydney",
//			expected:   "1985-01-01 22:00:00 AEST",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			assert.NoError(t, err)
//			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
//		})
//	}
//}

func TestTimezoneExoticLocations(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Kiritimati (Christmas Island, UTC+14)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Pacific/Kiritimati",
			expected:   "2024-01-16 02:00:00 LINT",
		},
		{
			name:       "Baker Island (UTC-12)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Etc/GMT+12",
			expected:   "2024-01-15 00:00:00 GMT+12",
		},
		{
			name:       "Tonga (UTC+13)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Pacific/Tongatapu",
			expected:   "2024-01-16 01:00:00 TOT",
		},
		{
			name:       "Marquesas Islands (UTC-9:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Pacific/Marquesas",
			expected:   "2024-01-15 02:30:00 MART",
		},
		{
			name:       "Cocos Islands (UTC+6:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Indian/Cocos",
			expected:   "2024-01-15 18:30:00 CCT",
		},
		{
			name:       "Myanmar (UTC+6:30)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Yangon",
			expected:   "2024-01-15 18:30:00 MMT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

// valid
func TestTimezoneMilitaryTimezones(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Zulu Time (Z/UTC)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Z",
			expected:   "2024-01-15 12:00:00 Z",
		},
		{
			name:       "Alpha Time Zone (A/UTC+1)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "A",
			expected:   "2024-01-15 13:00:00 A",
		},
		{
			name:       "Mike Time Zone (M/UTC+12)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "M",
			expected:   "2024-01-16 00:00:00 M",
		},
		{
			name:       "Yankee Time Zone (Y/UTC-12)",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Y",
			expected:   "2024-01-15 00:00:00 Y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

//func TestTimezoneHistoricalChanges(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expected   string
//	}{
//		{
//			name:       "China pre-1949",
//			sourceTime: "1940-01-15 12:00:00 UTC",
//			targetZone: "Asia/Shanghai",
//			expected:   "1940-01-15 20:00:00 CST",
//		},
//		{
//			name:       "India pre-1947",
//			sourceTime: "1940-01-15 12:00:00 UTC",
//			targetZone: "Asia/Kolkata",
//			expected:   "1940-01-15 17:30:00 IST",
//		},
//		{
//			name:       "Alaska Purchase (1867)",
//			sourceTime: "1867-10-18 12:00:00 UTC",
//			targetZone: "America/Juneau",
//			expected:   "1867-10-18 03:00:00 LMT",
//		},
//		{
//			name:       "Singapore pre-1982",
//			sourceTime: "1981-12-31 12:00:00 UTC",
//			targetZone: "Asia/Singapore",
//			expected:   "1981-12-31 19:30:00 SGT",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			assert.NoError(t, err)
//			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
//		})
//	}
//}

func TestTimezoneBoundaryConditions(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Year 2038 Problem",
			sourceTime: "2038-01-19 03:14:07 UTC",
			targetZone: "America/New_York",
			expected:   "2038-01-18 22:14:07 EST",
		},
		{
			name:       "Year 1900 Non-Leap Year",
			sourceTime: "1900-02-28 23:59:59 UTC",
			targetZone: "Europe/London",
			expected:   "1900-02-28 23:59:59 GMT",
		},
		{
			name:       "Year 2100 Non-Leap Year",
			sourceTime: "2100-02-28 23:59:59 UTC",
			targetZone: "Europe/Paris",
			expected:   "2100-03-01 00:59:59 CET",
		},
		{
			name:       "Unix Epoch Start",
			sourceTime: "1970-01-01 00:00:00 UTC",
			targetZone: "America/Los_Angeles",
			expected:   "1969-12-31 16:00:00 PST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezoneDSTEdgeCases(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Ambiguous Fall Back Time",
			sourceTime: "2024-11-03 01:30:00 EDT",
			targetZone: "America/New_York",
			expected:   "2024-11-03 01:30:00 EDT",
		},
		{
			name:       "Spring Forward Skip",
			sourceTime: "2024-03-10 02:30:00 EST",
			targetZone: "America/New_York",
			expected:   "2024-03-10 03:30:00 EDT",
		},
		{
			name:       "Southern Hemisphere DST Start",
			sourceTime: "2024-09-29 01:59:59 AEST",
			targetZone: "Australia/Sydney",
			expected:   "2024-09-29 01:59:59 AEST",
		},
		{
			name:       "Israel Variable DST",
			sourceTime: "2024-03-29 02:00:00 IST",
			targetZone: "Asia/Jerusalem",
			expected:   "2024-03-29 02:00:00 IST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

// valid
func BenchmarkTimeConversion(b *testing.B) {
	conv := setupConverter(nil)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
	}{
		{"Simple", "2024-01-15 12:00:00 UTC", "EST"},
		{"Complex", "2024-03-10 02:30:00 EST", "Australia/Lord_Howe"},
		{"Historical", "1940-01-15 12:00:00 UTC", "Asia/Shanghai"},
		{"DST Transition", "2024-03-10 02:00:00 EST", "America/New_York"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// valid
func TestTimezoneParallelConversion(t *testing.T) {
	conv := setupConverter(t)
	t.Run("Parallel", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 100; i++ {
			sourceTime := fmt.Sprintf("2024-01-15 %02d:00:00 UTC", i%24)
			_, err := conv.ConvertTimeZone(sourceTime, "Asia/Tokyo")
			assert.NoError(t, err)
		}
	})
}
func TestTimezoneSpecificRegionalCases(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		// Indiana Counties
		{
			name:       "Indiana - Knox County",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Knox",
			expected:   "2024-01-15 06:00:00 CST",
		},
		{
			name:       "Indiana - Marengo",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Marengo",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Indiana - Petersburg",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Petersburg",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Indiana - Tell City",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Tell_City",
			expected:   "2024-01-15 06:00:00 CST",
		},
		{
			name:       "Indiana - Vevay",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Vevay",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Indiana - Vincennes",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Vincennes",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Indiana - Winamac",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "America/Indiana/Winamac",
			expected:   "2024-01-15 07:00:00 EST",
		},

		// Arizona Regions
		{
			name:       "Arizona - Phoenix (No DST)",
			sourceTime: "2024-07-15 12:00:00 UTC",
			targetZone: "America/Phoenix",
			expected:   "2024-07-15 05:00:00 MST",
		},
		{
			name:       "Arizona - Navajo Nation (Uses DST)",
			sourceTime: "2024-07-15 12:00:00 UTC",
			targetZone: "America/Shiprock",
			expected:   "2024-07-15 06:00:00 MDT",
		},

		// Antarctica Research Stations
		{
			name:       "Antarctica - McMurdo",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/McMurdo",
			expected:   "2024-01-16 01:00:00 NZDT",
		},
		{
			name:       "Antarctica - Casey",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/Casey",
			expected:   "2024-01-15 19:00:00 +07",
		},
		{
			name:       "Antarctica - Davis",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/Davis",
			expected:   "2024-01-15 19:00:00 +07",
		},
		{
			name:       "Antarctica - DumontDUrville",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/DumontDUrville",
			expected:   "2024-01-15 22:00:00 +10",
		},
		{
			name:       "Antarctica - Syowa",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/Syowa",
			expected:   "2024-01-15 15:00:00 +03",
		},
		{
			name:       "Antarctica - Troll",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Antarctica/Troll",
			expected:   "2024-01-15 13:00:00 +01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

//func TestTimezoneLeapSecondHandling(t *testing.T) {
//	conv := setupConverter(t)
//	tests := []struct {
//		name       string
//		sourceTime string
//		targetZone string
//		expected   string
//	}{
//		{
//			name:       "Leap Second 2016",
//			sourceTime: "2016-12-31 23:59:60 UTC",
//			targetZone: "America/New_York",
//			expected:   "2016-12-31 18:59:60 EST",
//		},
//		{
//			name:       "Post Leap Second 2016",
//			sourceTime: "2017-01-01 00:00:00 UTC",
//			targetZone: "America/New_York",
//			expected:   "2016-12-31 19:00:00 EST",
//		},
//		{
//			name:       "Leap Second 2012",
//			sourceTime: "2012-06-30 23:59:60 UTC",
//			targetZone: "Asia/Tokyo",
//			expected:   "2012-07-01 08:59:60 JST",
//		},
//		{
//			name:       "Post Leap Second 2012",
//			sourceTime: "2012-07-01 00:00:00 UTC",
//			targetZone: "Asia/Tokyo",
//			expected:   "2012-07-01 09:00:00 JST",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
//			assert.NoError(t, err)
//			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
//		})
//	}
//}

func TestTimezoneFormatVariations(t *testing.T) {
	conv := setupConverter(t)
	sourceTime := "2024-01-15 12:00:00 UTC"
	targetZone := "America/New_York"

	formats := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "RFC3339",
			format:   time.RFC3339,
			expected: "2024-01-15T07:00:00-05:00",
		},
		{
			name:     "RFC822",
			format:   time.RFC822,
			expected: "15 Jan 24 07:00 EST",
		},
		{
			name:     "RFC850",
			format:   time.RFC850,
			expected: "Monday, 15-Jan-24 07:00:00 EST",
		},
		{
			name:     "RFC1123",
			format:   time.RFC1123,
			expected: "Mon, 15 Jan 2024 07:00:00 EST",
		},
		{
			name:     "Kitchen",
			format:   time.Kitchen,
			expected: "7:00AM",
		},
		{
			name:     "ANSIC",
			format:   time.ANSIC,
			expected: "Mon Jan 15 07:00:00 2024",
		},
		{
			name:     "UnixDate",
			format:   time.UnixDate,
			expected: "Mon Jan 15 07:00:00 EST 2024",
		},
		{
			name:     "RubyDate",
			format:   time.RubyDate,
			expected: "Mon Jan 15 07:00:00 -0500 2024",
		},
	}

	for _, tf := range formats {
		t.Run(tf.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(sourceTime, targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tf.expected, result.Format(tf.format))
		})
	}
}

func TestTimezoneTimezoneAbbreviationNormalization(t *testing.T) {
	conv := setupConverter(t)
	tests := []struct {
		name       string
		sourceTime string
		targetZone string
		expected   string
	}{
		{
			name:       "Mixed Case EST",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "eSt",
			expected:   "2024-01-15 07:00:00 EST",
		},
		{
			name:       "Lower Case pst",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "pst",
			expected:   "2024-01-15 04:00:00 PST",
		},
		{
			name:       "Mixed Case JsT",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "JsT",
			expected:   "2024-01-15 21:00:00 JST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertTimeZone(tt.sourceTime, tt.targetZone)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05 MST"))
		})
	}
}

func TestTimezonePropertyBasedTimezones(t *testing.T) {
	conv := setupConverter(t)

	// Test that converting to UTC and back preserves the original time
	t.Run("UTC Roundtrip", func(t *testing.T) {
		original := "2024-01-15 12:34:56 EST"
		intermediate, err := conv.ConvertTimeZone(original, "UTC")
		assert.NoError(t, err)

		result, err := conv.ConvertTimeZone(intermediate.Format("2006-01-02 15:04:05 MST"), "EST")
		assert.NoError(t, err)
		assert.Equal(t, original, result.Format("2006-01-02 15:04:05 MST"))
	})

	// Test that consecutive conversions are consistent
	t.Run("Conversion Chain", func(t *testing.T) {
		zones := []string{"UTC", "EST", "PST", "JST", "UTC"}
		testTime := "2024-01-15 12:00:00 UTC"
		var err error
		var result time.Time

		for i := 1; i < len(zones); i++ {
			result, err = conv.ConvertTimeZone(testTime, zones[i])
			assert.NoError(t, err)
		}

		// Should be back to original UTC time
		assert.Equal(t, "2024-01-15 12:00:00 UTC", result)
	})
}

// valid
func BenchmarkComplexScenarios(b *testing.B) {
	conv := setupConverter(nil)
	scenarios := []struct {
		name       string
		sourceTime string
		targetZone string
	}{
		{
			name:       "DST Transition",
			sourceTime: "2024-03-10 02:00:00 EST",
			targetZone: "America/New_York",
		},
		{
			name:       "Leap Second",
			sourceTime: "2016-12-31 23:59:59 UTC",
			targetZone: "Asia/Tokyo",
		},
		{
			name:       "Historical Date",
			sourceTime: "1900-01-01 12:00:00 GMT",
			targetZone: "America/New_York",
		},
		{
			name:       "Quarter Hour Offset",
			sourceTime: "2024-01-15 12:00:00 UTC",
			targetZone: "Asia/Kathmandu",
		},
	}

	for _, s := range scenarios {
		b.Run(s.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := conv.ConvertTimeZone(s.sourceTime, s.targetZone)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
