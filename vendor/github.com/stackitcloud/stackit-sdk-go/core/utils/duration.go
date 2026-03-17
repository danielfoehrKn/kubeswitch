// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 STACKIT GmbH & Co. KG

package utils

// Package utils provides general utility functions for common operations.
//
// The package includes utilities for:
//   - Duration parsing and conversion with flexible unit support
//   - Time-based calculations including calendar-aware month handling
//   - Configurable validation with minimum/maximum boundaries
//
// Duration Conversion:
//
// The main function ConvertToSeconds parses time strings like "30m", "2h", "7d"
// and converts them to seconds. It supports both fixed-rate conversions and
// calendar-aware calculations for months.
//
// Supported units:
//   - "s": seconds
//   - "m": minutes
//   - "h": hours
//   - "d": days (24 hours)
//   - "M": months (calendar-aware, handles varying month lengths)
//
// Example usage:
//
//	// Basic conversion
//	seconds, err := utils.ConvertToSeconds("30m")
//
//	// With validation boundaries
//	seconds, err := utils.ConvertToSeconds("2h",
//		utils.WithMinSeconds(60),
//		utils.WithMaxSeconds(7200))
//
//	// Calendar-aware month calculation
//	seconds, err := utils.ConvertToSeconds("3M", utils.WithNow(time.Now()))
//
// The package uses the functional options pattern for flexible configuration
// and provides extensible interfaces for adding custom duration converters.

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// =============================================================================
// Public Types and Interfaces
// =============================================================================

// DurationConverter defines the interface for converting a numeric value into a
// time.Duration. This abstraction supports both fixed-rate conversions (like
// seconds to minutes) and complex calendar-aware calculations (like adding months).
type DurationConverter interface {
	// ToDuration converts a numeric value into a time.Duration.
	//
	// The function takes the following parameters:
	//   - value: The numeric value to convert (e.g., 30 for 30 days).
	//   - now: The reference time for calendar-based calculations. It can be
	//     ignored by implementations that are not calendar-dependent.
	//
	// It returns the calculated time.Duration and a nil error on success. On
	// failure, it returns an error, for example if the value is too large to
	// be processed.
	ToDuration(value uint64, now time.Time) (time.Duration, error)
}

// ValidationError represents errors that occur during input validation.
// It uses a Type field to distinguish between different validation failures.
type ValidationError struct {
	Type    ValidationErrorType // The specific type of validation error
	Input   string              // The invalid input value
	Reason  string              // Human-readable reason for the error
	Context map[string]any      // Additional context (min/max values, valid units, etc.)
}

type ValidationErrorType string

const (
	ValidationErrorInvalidFormat ValidationErrorType = "invalid_format"
	ValidationErrorInvalidValue  ValidationErrorType = "invalid_value"
	ValidationErrorInvalidUnit   ValidationErrorType = "invalid_unit"
	ValidationErrorBelowMinimum  ValidationErrorType = "below_minimum"
	ValidationErrorAboveMaximum  ValidationErrorType = "above_maximum"
)

func (e *ValidationError) Error() string {
	switch e.Type {
	case ValidationErrorInvalidFormat:
		if e.Reason != "" && e.Input != "" {
			return fmt.Sprintf("invalid time string format %q: %s", e.Input, e.Reason)
		}
		if e.Input != "" {
			return fmt.Sprintf("invalid time string format: %q", e.Input)
		}
		return "invalid time string format"

	case ValidationErrorInvalidValue:
		if e.Reason != "" && e.Input != "" {
			return fmt.Sprintf("invalid time value %q: %s", e.Input, e.Reason)
		}
		if e.Input != "" {
			return fmt.Sprintf("invalid time value: %q", e.Input)
		}
		return "invalid time value"

	case ValidationErrorInvalidUnit:
		if e.Input != "" {
			if validUnits, ok := e.Context["validUnits"].([]string); ok {
				return fmt.Sprintf("invalid time unit %q, supported units are %v", e.Input, validUnits)
			}
			return fmt.Sprintf("invalid time unit: %q", e.Input)
		}
		return "invalid time unit"

	case ValidationErrorBelowMinimum:
		if minValue, ok := e.Context["minimum"].(uint64); ok {
			if val, ok := e.Context["value"].(uint64); ok {
				return fmt.Sprintf("duration is below minimum: %d seconds (minimum: %d seconds)", val, minValue)
			}
		}
		return "duration is below the allowed minimum"

	case ValidationErrorAboveMaximum:
		if maxValue, ok := e.Context["maximum"].(uint64); ok {
			if val, ok := e.Context["value"].(uint64); ok {
				return fmt.Sprintf("duration exceeds maximum: %d seconds (maximum: %d seconds)", val, maxValue)
			}
		}
		return "duration exceeds the allowed maximum"

	default:
		return fmt.Sprintf("validation error: %s", e.Input)
	}
}

// Is implements error matching for errors.Is()
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return e.Type == t.Type
	}
	return false
}

// CalculationError represents errors that occur during duration calculations.
type CalculationError struct {
	Type    CalculationErrorType // The specific type of calculation error
	Value   uint64               // The value that caused the error
	Reason  string               // Human-readable reason
	Context map[string]any       // Additional context (multiplier, operation, etc.)
}

type CalculationErrorType string

const (
	CalculationErrorOutOfBounds        CalculationErrorType = "out_of_bounds"
	CalculationErrorNegativeResult     CalculationErrorType = "negative_result"
	CalculationErrorNegativeMultiplier CalculationErrorType = "negative_multiplier"
)

func (e *CalculationError) Error() string {
	switch e.Type {
	case CalculationErrorOutOfBounds:
		msg := "calculation result is out of bounds"
		if e.Value > 0 {
			msg += fmt.Sprintf(" (value: %d)", e.Value)
		}
		if operation, ok := e.Context["operation"].(string); ok {
			msg += fmt.Sprintf(" during %s", operation)
		}
		if limit, ok := e.Context["limit"].(string); ok {
			msg += fmt.Sprintf(" (exceeds %s)", limit)
		}
		return msg

	case CalculationErrorNegativeResult:
		if result, ok := e.Context["result"].(float64); ok {
			return fmt.Sprintf("calculated duration is negative: %f", result)
		}
		return "calculated duration is negative"

	case CalculationErrorNegativeMultiplier:
		if multiplier, ok := e.Context["multiplier"].(time.Duration); ok {
			return fmt.Sprintf("duration multiplier is negative: %v", multiplier)
		}
		return "duration multiplier is negative"

	default:
		return fmt.Sprintf("calculation error with value %d", e.Value)
	}
}

// Is implements error matching for errors.Is()
func (e *CalculationError) Is(target error) bool {
	if t, ok := target.(*CalculationError); ok {
		return e.Type == t.Type
	}
	return false
}

// =============================================================================
// Public Functions
// =============================================================================
// NewValidationError creates a new ValidationError with the specified type and details.
func NewValidationError(errorType ValidationErrorType, input, reason string, context map[string]any) *ValidationError {
	return &ValidationError{
		Type:    errorType,
		Input:   input,
		Reason:  reason,
		Context: context,
	}
}

// NewCalculationError creates a new CalculationError with the specified type and details.
func NewCalculationError(errorType CalculationErrorType, value uint64, reason string, context map[string]any) *CalculationError {
	return &CalculationError{
		Type:    errorType,
		Value:   value,
		Reason:  reason,
		Context: context,
	}
}

// Constructors for common error cases
func NewInvalidFormatError(input, reason string) *ValidationError {
	return NewValidationError(ValidationErrorInvalidFormat, input, reason, nil)
}

func NewInvalidValueError(input, reason string) *ValidationError {
	return NewValidationError(ValidationErrorInvalidValue, input, reason, nil)
}

func NewInvalidUnitError(unit string, validUnits []string) *ValidationError {
	context := map[string]any{"validUnits": validUnits}
	return NewValidationError(ValidationErrorInvalidUnit, unit, "", context)
}

func NewBelowMinimumError(value, minimum uint64) *ValidationError {
	context := map[string]any{"value": value, "minimum": minimum}
	return NewValidationError(ValidationErrorBelowMinimum, "", "", context)
}

func NewAboveMaximumError(value, maximum uint64) *ValidationError {
	context := map[string]any{"value": value, "maximum": maximum}
	return NewValidationError(ValidationErrorAboveMaximum, "", "", context)
}

// An Option configures a ConvertToSeconds call.
type Option func(*converterConfig)

// WithMinSeconds sets a minimum duration in seconds. A value of 0 means no limit.
func WithMinSeconds(minSeconds uint64) Option {
	return func(c *converterConfig) {
		c.minSeconds = &minSeconds
	}
}

// WithMaxSeconds sets a maximum duration in seconds. A value of 0 means no limit.
func WithMaxSeconds(maxSeconds uint64) Option {
	return func(c *converterConfig) {
		if maxSeconds == 0 {
			c.maxSeconds = nil // Remove any previously set limit when 0
		} else {
			c.maxSeconds = &maxSeconds
		}
	}
}

// WithNow sets the current time to be used by the DurationConverter interface's ToDuration function.
// Useful for determenistic when testing calendar-dependent implementations of DurationConverter.
func WithNow(now time.Time) Option {
	return func(c *converterConfig) {
		c.now = &now
	}
}

// WithUnits provides a custom map of duration converters.
func WithUnits(units map[string]DurationConverter) Option {
	return func(c *converterConfig) {
		c.units = units
	}
}

// ConvertToSeconds converts a time string in the form of <value><unit> (e.g., "30m", "1h") into a total number of seconds as a uint64.
//
// The function uses a default set of units:
//   - "s": seconds
//   - "m": minutes
//   - "h": hours
//   - "d": days 	(24 hours)
//   - "M": months 	(calendar-aware)
//
// Optional configurations can be provided using the Option type, such as setting
// minimum/maximum second boundaries (WithMinSeconds, WithMaxSeconds) or providing
// a custom set of units (WithUnits).
//
// It returns an error if the format is invalid, the unit is unsupported, or if the
// calculated value violates any provided boundaries.
func ConvertToSeconds(timeStr string, opts ...Option) (uint64, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Define a default config
	cfg := &converterConfig{
		units: defaultUnits,
	}
	// Apply optional config settings, overwriting defaults
	for _, opt := range opts {
		opt(cfg)
	}

	// Separate unit and value of timeStr
	valueStr, unit, err := splitValueAndUnit(timeStr)
	if err != nil {
		return 0, err
	}

	// Check for leading zeros, which are not allowed for values > 0
	if len(valueStr) > 1 && valueStr[0] == '0' {
		return 0, NewInvalidFormatError(timeStr, "leading zeros are not allowed")
	}

	// Parse Value, make sure it's a valid positive number that fit's a uint64
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return 0, NewInvalidValueError(valueStr, err.Error())
	}

	// A value of 0 is not a valid duration.
	if value == 0 {
		return 0, NewInvalidValueError("0", "a value of 0 is not allowed")
	}

	// Look up the DurationConverter to use
	converter, ok := cfg.units[unit]
	if !ok {
		return 0, NewInvalidUnitError(unit, getKeys(cfg.units))
	}

	// Use the provided time.Time for 'now' if provided. Otherwise use system time.
	var now time.Time
	if cfg.now != nil {
		now = *cfg.now
	} else {
		now = time.Now()
	}

	// Calculate time.Duration using the DurationConverter interface
	totalDuration, err := converter.ToDuration(value, now)
	if err != nil {
		return 0, err
	}
	// Convert time.Duration to Seconds
	secondsFloat := totalDuration.Seconds() // float64

	// Check for negative or overflow values and return an error if necessary
	if secondsFloat < 0 {
		context := map[string]any{"result": secondsFloat}
		return 0, NewCalculationError(CalculationErrorNegativeResult, 0, "", context)
	}
	if secondsFloat > math.MaxUint64 {
		// This case can only be tiggered if a new positive integer type bigger than MaxUint64 is added to Go
		// because we currently can not have a custom converter returning bigger values than MaxUint64.
		// Thus we currently can not fully test this path but will leave it here (defensive programming).
		context := map[string]any{
			"operation": "final conversion",
			"limit":     "MaxUint64",
		}
		return 0, NewCalculationError(CalculationErrorOutOfBounds, 0, "result exceeds MaxUint64", context)
	}

	// Cast to uint64 for boundary checks and final return
	seconds := uint64(secondsFloat)

	// Check if the calculated duration is within the specified, optional boundaries.
	if cfg.minSeconds != nil && seconds < *cfg.minSeconds {
		return 0, NewBelowMinimumError(seconds, *cfg.minSeconds)
	}
	if cfg.maxSeconds != nil && seconds > *cfg.maxSeconds {
		return 0, NewAboveMaximumError(seconds, *cfg.maxSeconds)
	}

	return seconds, nil
}

// =============================================================================
// Private Types and Interfaces
// =============================================================================

// FixedMultiplier converts a value to a duration using a constant multiplier.
// It implements the DurationConverter interface.
type fixedMultiplier struct {
	Multiplier time.Duration
}

// fixedMultiplier.ToDuration calculates the duration by multiplying the value with the fixed multiplier.
// The `now` parameter is ignored as this calculation is not calendar-dependent.
func (fm fixedMultiplier) ToDuration(value uint64, _ time.Time) (time.Duration, error) {
	// A negative multiplier is invalid as it would produce a negative duration and break the following overflow check. Fail early.
	if fm.Multiplier < 0 {
		context := map[string]any{"multiplier": fm.Multiplier}
		return 0, NewCalculationError(CalculationErrorNegativeMultiplier, 0, "", context)
	}

	// A zero multiplier will always result in a zero duration.
	if fm.Multiplier == 0 {
		return 0, nil
	}

	// Check for overflow: both the value AND the multiplication result must fit in int64
	multiplierUint := uint64(fm.Multiplier) // #nosec G115 - already validated fm.Multiplier >= 0
	if value > math.MaxInt64 || value > math.MaxInt64/multiplierUint {
		context := map[string]any{
			"operation": "multiplication",
			"limit":     "MaxInt64",
		}
		return 0, NewCalculationError(CalculationErrorOutOfBounds, value, "value or multiplication result exceeds MaxInt64", context)
	}

	return time.Duration(int64(value)) * fm.Multiplier, nil
}

// MonthCalculator implements calendar-aware month addition.
// Note: Results are normalized - adding 1 month to January 31st
// results in March 2nd/3rd (February 31st normalized).
// This matches the behavior of Go's time.AddDate().
type monthCalculator struct{}

// monthCalculator.ToDuration calculates the time.Duration between now and a future date N months from now.
func (mc monthCalculator) ToDuration(value uint64, now time.Time) (time.Duration, error) {
	// Check if casting to int for AddDate would cause an overflow.
	if value > math.MaxInt {
		context := map[string]any{
			"operation": "month calculation",
			"limit":     "MaxInt",
		}
		return 0, NewCalculationError(CalculationErrorOutOfBounds, value, "value exceeds MaxInt", context)
	}
	future := now.AddDate(0, int(value), 0)
	return future.Sub(now), nil
}

// config holds the optional parameters for ConvertToSeconds.
type converterConfig struct {
	minSeconds *uint64
	maxSeconds *uint64
	units      map[string]DurationConverter
	now        *time.Time
}

// =============================================================================
// Private Variables and Constants
// =============================================================================

// DefaultUnits is a map of supported unit characters to their multipliers.
var defaultUnits = map[string]DurationConverter{
	"s": fixedMultiplier{Multiplier: time.Second},
	"m": fixedMultiplier{Multiplier: time.Minute},
	"h": fixedMultiplier{Multiplier: time.Hour},
	"d": fixedMultiplier{Multiplier: 24 * time.Hour}, // Day
	"M": monthCalculator{},                           // Month (calendar-aware)
}

// =============================================================================
// Private Functions
// =============================================================================

// Helper to get all keys of DurationConverter for error logging purposes.
func getKeys(m map[string]DurationConverter) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// splitValueAndUnit separates a time string like "30m" into its numeric ("30") and unit ("m") parts. It handles multi-character units.
func splitValueAndUnit(timeStr string) (valueStr, unitStr string, err error) {
	if timeStr == "" {
		return "", "", NewInvalidFormatError("", "input string is empty")
	}

	// Find the first non-digit character in the string.
	var splitIndex = -1
	for i, r := range timeStr {
		if !unicode.IsDigit(r) {
			// If the first non-digit is a '.' or ',' it's a float, which we don't support.
			if r == '.' || r == ',' {
				return "", "", NewInvalidValueError(timeStr, "floating-point values are not supported")
			}
			splitIndex = i
			break
		}
	}

	// Check for invalid formats
	if splitIndex == -1 {
		return "", "", NewInvalidFormatError(timeStr, "contains no unit, expected format <value><unit>")
	}
	if splitIndex == 0 {
		return "", "", NewInvalidFormatError(timeStr, "must start with a number, expected format <value><unit>")
	}

	valueStr = timeStr[:splitIndex]
	unitStr = timeStr[splitIndex:]
	return valueStr, unitStr, nil
}
