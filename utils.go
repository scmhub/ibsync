package ibsync

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TickTypesToString converts a variadic number of TickType values to a comma-separated string representation.
func TickTypesToString(tt ...TickType) string {
	var strs []string
	for _, t := range tt {
		strs = append(strs, fmt.Sprintf("%d", t))
	}
	return strings.Join(strs, ",")
}

// isDigit checks if the provided string consists only of digits.
func isDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// FormatIBTime formats a time.Time value into the string format required by IB API.
//
// The returned string is in the format "YYYYMMDD HH:MM:SS", which is used by Interactive Brokers for specifying date and time.
// As no time zone is provided IB will use your local time zone so we enforce local time zone first.
// It returns "" if time is zero.
func FormatIBTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format("20060102 15:04:05")
}

// FormatIBTimeUTC sets a time.Time value to UTC time zone  and formats into the string format required by IB API.
//
// Note that there is a dash between the date and time in UTC notation.
// The returned string is in the format "YYYYMMDD HH:MM:SS UTC", which is used by Interactive Brokers for specifying date and time.
// It returns "" if time is zero.
func FormatIBTimeUTC(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("20060102-15:04:05") + " UTC"
}

// FormatIBTimeUSEastern sets a time.Time value to US/Eastern time zone  and formats into the string format required by IB API.
//
// The returned string is in the format "YYYYMMDD HH:MM:SS US/Eastern", which is used by Interactive Brokers for specifying date and time.
// It returns "" if time is zero.
func FormatIBTimeUSEastern(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	loc, _ := time.LoadLocation("America/New_York")
	return t.In(loc).Format("20060102 15:04:05") + " US/Eastern"
}

// ParseIBTime parses an IB string representation of time into a time.Time object.
// It supports various IB formats, including:
// 1. "YYYYMMDD"
// 2. Unix timestamp ("1617206400")
// 3. "YYYYMMDD HH:MM:SS Timezone"
// 4. // "YYYYmmdd  HH:MM:SS", "YYYY-mm-dd HH:MM:SS.0" or "YYYYmmdd-HH:MM:SS"
func ParseIBTime(s string) (time.Time, error) {
	var layout string
	// "YYYYMMDD"
	if len(s) == 8 {
		layout = "20060102"
		return time.Parse(layout, s)
	}
	// "1617206400"
	if isDigit(s) {
		ts, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(ts, 0), nil
	}
	// "20221125 10:00:00 Europe/Amsterdam"
	if strings.Count(s, " ") >= 2 && !strings.Contains(s, "  ") {
		split := strings.Split(s, " ")
		layout = "20060102 15:04:05"
		t, err := time.Parse(layout, split[0]+" "+split[1])
		if err != nil {
			return time.Time{}, err
		}
		loc, err := time.LoadLocation(split[2])
		if err != nil {
			return time.Time{}, err
		}
		return t.In(loc), nil
	}
	// "YYYYmmdd  HH:MM:SS", "YYYY-mm-dd HH:MM:SS.0" or "YYYYmmdd-HH:MM:SS"
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "  ", "")
	s = strings.ReplaceAll(s, " ", "")
	if len(s) > 15 {
		s = s[:16]
	}
	layout = "2006010215:04:05"
	return time.Parse(layout, s)
}

// LastWednesday12EST returns time.Time correspondong to last wednesday 12:00 EST
// Without checking holidays this a high probability time for open market. Usefull for testing historical data
func LastWednesday12EST() time.Time {
	// last Wednesday
	offset := (int(time.Now().Weekday()) - int(time.Wednesday) + 7) % 7
	if offset == 0 {
		offset = 7
	}
	lw := time.Now().AddDate(0, 0, -offset)
	EST, _ := time.LoadLocation("America/New_York")
	return time.Date(lw.Year(), lw.Month(), lw.Day(), 12, 0, 0, 0, EST)
}

// UpdateStruct copies non-zero fields from src to dest.
// dest must be a pointer to a struct so that it can be updated;
// src can be either a struct or a pointer to a struct.
func UpdateStruct(dest, src any) error {
	destVal := reflect.ValueOf(dest)
	srcVal := reflect.Indirect(reflect.ValueOf(src)) // Handles both struct and *struct for src

	// Ensure that dest is a pointer to a struct
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a struct")
	}
	// Ensure src is a struct (after dereferencing if it's a pointer)
	if srcVal.Kind() != reflect.Struct {
		return errors.New("src must be a struct or a pointer to a struct")
	}

	destVal = destVal.Elem() // Dereference dest to the actual struct
	srcType := srcVal.Type()

	// Iterate over each field in src struct
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		fieldName := srcType.Field(i).Name
		destField := destVal.FieldByName(fieldName)

		// Update if field exists in dest and src field is non-zero
		if destField.IsValid() && destField.CanSet() && !srcField.IsZero() {
			destField.Set(srcField)
		}
	}
	return nil
}

// Stringify converts a struct or pointer to a struct into a string representation
// Skipping fields with zero or nil values and dereferencing pointers.
// It is recursive and will apply to nested structs.
// It can NOT handle unexported fields.
func Stringify(obj interface{}) string {
	return stringifyValue(reflect.ValueOf(obj))
}

// isEmptyValue checks if a value is considered "empty"
func isEmptyValue(v reflect.Value) bool {
	// Handle nil pointers
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}

	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Special handling for Time struct
		if v.Type().String() == "time.Time" {
			return v.Interface().(time.Time).IsZero()
		}

		// Check if all fields are empty
		for i := 0; i < v.NumField(); i++ {
			if !isEmptyValue(v.Field(i)) {
				return false
			}
		}
		return true

	case reflect.Slice, reflect.Map:
		return v.Len() == 0

	case reflect.String:
		return v.String() == ""

	case reflect.Bool:
		return !v.Bool()

	case reflect.Int, reflect.Int32, reflect.Int64:
		return v.Int() == 0 || v.Int() == UNSET_INT || v.Int() == UNSET_LONG

	case reflect.Float32, reflect.Float64:
		return v.Float() == 0 || v.Float() == UNSET_FLOAT

	case reflect.Ptr:
		return v.IsNil()

	case reflect.Interface:
		return v.IsNil()
	}

	return false
}

// stringifyValue handles the recursive stringification of values
func stringifyValue(v reflect.Value) string {
	// Handle pointer types by dereferencing
	if v.Kind() == reflect.Ptr {
		// If nil, return empty string
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return stringifyStruct(v)
	case reflect.Slice:
		return stringifySlice(v)
	case reflect.Map:
		return stringifyMap(v)
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// stringifyStruct handles struct-specific stringification
func stringifyStruct(v reflect.Value) string {
	// Get the type of the struct
	t := v.Type()

	// Skip completely empty structs
	if isEmptyValue(v) {
		return ""
	}

	// Build the string representation
	var fields []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		// Skip unexported (private) fields
		if !field.CanInterface() {
			continue
		}

		// Skip empty values
		if isEmptyValue(field) {
			continue
		}

		// Recursively get the stringified value
		fieldValueStr := stringifyValue(field)

		// Add to fields if not empty
		if fieldValueStr != "" {
			fields = append(fields, fmt.Sprintf("%s=%s", fieldName, fieldValueStr))
		}
	}

	// Construct the final string
	if len(fields) == 0 {
		return ""
	}
	return fmt.Sprintf("%s{%s}", t.Name(), strings.Join(fields, ", "))
}

// stringifySlice handles slice stringification
func stringifySlice(v reflect.Value) string {
	// If slice is empty or nil
	if v.Len() == 0 {
		return ""
	}

	// Convert slice elements to strings
	var elements []string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// Skip empty values
		if isEmptyValue(elem) {
			continue
		}

		elemStr := stringifyValue(elem)
		if elemStr != "" {
			elements = append(elements, elemStr)
		}
	}

	// If no non-zero elements, return empty
	if len(elements) == 0 {
		return ""
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// stringifyMap handles map stringification
func stringifyMap(v reflect.Value) string {
	// If map is empty or nil
	if v.Len() == 0 {
		return ""
	}

	// Convert map elements to strings
	var elements []string
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		val := iter.Value()

		// Skip empty values
		if isEmptyValue(val) {
			continue
		}

		valStr := stringifyValue(val)
		if valStr != "" {
			elements = append(elements, fmt.Sprintf("%v=%s", k.Interface(), valStr))
		}
	}

	// If no non-zero elements, return empty
	if len(elements) == 0 {
		return ""
	}

	return fmt.Sprintf("map[%s]", strings.Join(elements, ", "))
}
