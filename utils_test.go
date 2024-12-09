package ibsync

import (
	"reflect"
	"testing"
	"time"
)

func TestIsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "123456", // All digits
			expected: true,
		},
		{
			input:    "001234", // Leading zeros, all digits
			expected: true,
		},
		{
			input:    "123a456", // Contains a non-digit character
			expected: false,
		},
		{
			input:    "abc", // All non-digit characters
			expected: false,
		},
		{
			input:    "",   // Empty string
			expected: true, // Edge case, no characters means no non-digits
		},
		{
			input:    "   ", // Spaces are not digits
			expected: false,
		},
		{
			input:    "123 456", // Contains a space
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isDigit(tt.input)
			if result != tt.expected {
				t.Errorf("For input '%s': expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestFormatIBTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "",
		},
		{
			name:     "typical local date time",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.Local),
			expected: "20240315 14:30:45",
		},
		{
			name:     "UTC time with nanoseconds",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC).In(time.Local).Format("20060102 15:04:05"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatIBTime(tt.input)
			if got != tt.expected {
				t.Errorf("FormatIBTime() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatIBTimeUTC(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "",
		},
		{
			name:     "typical date time",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
			expected: "20240315-14:30:45 UTC",
		},
		{
			name:     "convert from different timezone",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC).In(est),
			expected: "20240315-14:30:45 UTC", // Should convert back to UTC
		},
		{
			name:     "midnight UTC",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "20240101-00:00:00 UTC",
		},
		{
			name:     "with nanoseconds",
			input:    time.Date(2024, 1, 1, 12, 0, 0, 123456789, time.UTC),
			expected: "20240101-12:00:00 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatIBTimeUTC(tt.input)
			if got != tt.expected {
				t.Errorf("FormatIBTimeUTC() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatIBTimeUSEastern(t *testing.T) {
	EST, _ := time.LoadLocation("America/New_York")
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "",
		},
		{
			name:     "typical date time",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, EST),
			expected: "20240315 14:30:45 US/Eastern", // UTC-4 during DST
		},
		{
			name:     "UTC time conversion",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
			expected: "20240315 10:30:45 US/Eastern", // UTC-4 during DST
		},
		{
			name:     "during EST (non-DST)",
			input:    time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
			expected: "20240115 09:30:45 US/Eastern", // UTC-5 during EST
		},
		{
			name:     "DST transition spring forward",
			input:    time.Date(2024, 3, 10, 14, 30, 45, 0, time.UTC),
			expected: "20240310 10:30:45 US/Eastern",
		},
		{
			name:     "DST transition fall back",
			input:    time.Date(2024, 11, 3, 14, 30, 45, 0, time.UTC),
			expected: "20241103 09:30:45 US/Eastern",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatIBTimeUSEastern(tt.input)
			if got != tt.expected {
				t.Errorf("FormatIBTimeUSEastern() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseIBTime(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
		hasError bool
	}{
		{
			input:    "20231016", // YYYYMMDD
			expected: time.Date(2023, 10, 16, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			input:    "1617206400", // Unix timestamp
			expected: time.Unix(1617206400, 0),
			hasError: false,
		},
		{
			input:    "20221125 10:00:00 Europe/Amsterdam",                           // DateTime with timezone
			expected: time.Date(2022, 11, 25, 10, 0, 0, 0, time.FixedZone("CET", 0)), // Adjust to CET timezone
			hasError: false,
		},
		{
			input:    "2023-10-16  10:00:00", // YYYY-mm-dd  HH:MM:SS
			expected: time.Date(2023, 10, 16, 10, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			input:    "2023-10-16 10:00:00.0", // YYYY-mm-dd HH:MM:SS.0
			expected: time.Date(2023, 10, 16, 10, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			input:    "20231016-10:00:00", // YYYY-mm-dd-HH:MM:SS
			expected: time.Date(2023, 10, 16, 10, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			input:    "invalid-string", // Invalid format
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseIBTime(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected an error but got none for input: %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v for input: %s", err, tt.input)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("expected: %v, got: %v for input: %s", tt.expected, result, tt.input)
				}
			}
		})
	}
}

// Example struct for testing
type Example struct {
	Name    string
	Age     int
	Address string
	Active  bool
}

// Test function to test the UpdateStruct behavior
func TestUpdateStruct(t *testing.T) {
	tests := []struct {
		name     string
		dest     Example
		src      Example
		expected Example
	}{
		{
			name:     "Non-zero fields in src update dest",
			dest:     Example{Name: "Alice", Age: 25, Address: "Old Address", Active: false},
			src:      Example{Name: "Bob", Age: 30}, // Only Name and Age should update
			expected: Example{Name: "Bob", Age: 30, Address: "Old Address", Active: false},
		},
		{
			name:     "Zero fields in src do not update dest",
			dest:     Example{Name: "Alice", Age: 25, Address: "Old Address", Active: true},
			src:      Example{Address: ""}, // Empty string should not override
			expected: Example{Name: "Alice", Age: 25, Address: "Old Address", Active: true},
		},
		{
			name:     "Empty src does not change dest",
			dest:     Example{Name: "Alice", Age: 25, Address: "Old Address", Active: true},
			src:      Example{}, // No fields to update
			expected: Example{Name: "Alice", Age: 25, Address: "Old Address", Active: true},
		},
		{
			name:     "Update boolean field in dest",
			dest:     Example{Name: "Alice", Age: 25, Address: "Old Address", Active: false},
			src:      Example{Active: true}, // Only Active should update
			expected: Example{Name: "Alice", Age: 25, Address: "Old Address", Active: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of dest to pass as a pointer
			dest := tt.dest

			// Call UpdateStruct with a pointer to dest and src
			if err := UpdateStruct(&dest, tt.src); err != nil {
				t.Fatalf("UpdateStruct failed: %v", err)
			}

			// Check if dest matches the expected result
			if !reflect.DeepEqual(dest, tt.expected) {
				t.Errorf("UpdateStruct() = %v, want %v", dest, tt.expected)
			}
		})
	}
}

// Test for handling incorrect types in dest or src
func TestUpdateStructInvalidTypes(t *testing.T) {
	var dest Example
	src := Example{Name: "Bob"}

	// Non-pointer dest should return an error
	if err := UpdateStruct(dest, src); err == nil {
		t.Errorf("Expected error for non-pointer dest, got nil")
	}

	// dest as a pointer but src as a non-struct should return an error
	var nonStructSrc int
	if err := UpdateStruct(&dest, nonStructSrc); err == nil {
		t.Errorf("Expected error for non-struct src, got nil")
	}
}
