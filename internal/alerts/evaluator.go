package alerts

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pihole-analyzer/internal/logger"
)

// BasicEvaluator implements the Evaluator interface
type BasicEvaluator struct {
	logger *logger.Logger
}

// NewEvaluator creates a new condition evaluator
func NewEvaluator(logger *logger.Logger) Evaluator {
	return &BasicEvaluator{
		logger: logger.Component("alert-evaluator"),
	}
}

// EvaluateConditions evaluates all conditions with AND logic
func (e *BasicEvaluator) EvaluateConditions(conditions []AlertCondition, data map[string]interface{}) (bool, error) {
	if len(conditions) == 0 {
		return false, nil
	}

	for _, condition := range conditions {
		result, err := e.EvaluateCondition(condition, data)
		if err != nil {
			e.logger.GetSlogger().Error("Failed to evaluate condition",
				slog.String("field", condition.Field),
				slog.String("operator", condition.Operator),
				slog.String("error", err.Error()))
			return false, err
		}

		// AND logic - all conditions must be true
		if !result {
			return false, nil
		}
	}

	return true, nil
}

// EvaluateCondition evaluates a single condition against the data
func (e *BasicEvaluator) EvaluateCondition(condition AlertCondition, data map[string]interface{}) (bool, error) {
	// Get the field value from data
	fieldValue, exists := data[condition.Field]
	if !exists {
		e.logger.GetSlogger().Debug("Field not found in data", slog.String("field", condition.Field))
		return false, nil
	}

	// Convert expected value to appropriate type if needed (skip for array operations)
	var expectedValue interface{}
	if condition.Operator == "in" || condition.Operator == "not_in" {
		expectedValue = condition.Value // Don't convert for array operations
	} else {
		var err error
		expectedValue, err = e.convertValue(condition.Value, fieldValue)
		if err != nil {
			return false, fmt.Errorf("failed to convert expected value: %w", err)
		}
	}

	// Apply time window if specified
	if condition.TimeWindow != "" {
		windowData, err := e.applyTimeWindow(condition.Field, condition.TimeWindow, data)
		if err != nil {
			return false, fmt.Errorf("failed to apply time window: %w", err)
		}
		fieldValue = windowData
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "gt", ">":
		return e.compareNumbers(fieldValue, expectedValue, func(a, b float64) bool { return a > b })
	case "gte", ">=":
		return e.compareNumbers(fieldValue, expectedValue, func(a, b float64) bool { return a >= b })
	case "lt", "<":
		return e.compareNumbers(fieldValue, expectedValue, func(a, b float64) bool { return a < b })
	case "lte", "<=":
		return e.compareNumbers(fieldValue, expectedValue, func(a, b float64) bool { return a <= b })
	case "eq", "==", "=":
		return e.compareEqual(fieldValue, expectedValue)
	case "ne", "!=":
		result, err := e.compareEqual(fieldValue, expectedValue)
		return !result, err
	case "contains":
		return e.containsString(fieldValue, expectedValue)
	case "not_contains":
		result, err := e.containsString(fieldValue, expectedValue)
		return !result, err
	case "regex":
		return e.matchesRegex(fieldValue, expectedValue)
	case "in":
		return e.inArray(fieldValue, expectedValue)
	case "not_in":
		result, err := e.inArray(fieldValue, expectedValue)
		return !result, err
	default:
		return false, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

// convertValue converts the expected value to match the field value type
func (e *BasicEvaluator) convertValue(expected interface{}, fieldValue interface{}) (interface{}, error) {
	if expected == nil || fieldValue == nil {
		return expected, nil
	}

	expectedType := reflect.TypeOf(fieldValue)
	expectedValue := reflect.ValueOf(expected)

	// If types already match, return as-is
	if expectedValue.Type() == expectedType {
		return expected, nil
	}

	// Try to convert to the target type
	switch expectedType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.convertToInt(expected)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return e.convertToUint(expected)
	case reflect.Float32, reflect.Float64:
		return e.convertToFloat(expected)
	case reflect.String:
		return fmt.Sprintf("%v", expected), nil
	case reflect.Bool:
		return e.convertToBool(expected)
	default:
		return expected, nil
	}
}

// convertToInt converts a value to int64
func (e *BasicEvaluator) convertToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// convertToUint converts a value to uint64
func (e *BasicEvaluator) convertToUint(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case int:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", value)
	}
}

// convertToFloat converts a value to float64
func (e *BasicEvaluator) convertToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// convertToBool converts a value to bool
func (e *BasicEvaluator) convertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// compareNumbers compares two values as numbers
func (e *BasicEvaluator) compareNumbers(actual, expected interface{}, compare func(float64, float64) bool) (bool, error) {
	actualFloat, err := e.convertToFloat(actual)
	if err != nil {
		return false, fmt.Errorf("failed to convert actual value to float: %w", err)
	}

	expectedFloat, err := e.convertToFloat(expected)
	if err != nil {
		return false, fmt.Errorf("failed to convert expected value to float: %w", err)
	}

	return compare(actualFloat, expectedFloat), nil
}

// compareEqual compares two values for equality
func (e *BasicEvaluator) compareEqual(actual, expected interface{}) (bool, error) {
	// Handle nil cases
	if actual == nil && expected == nil {
		return true, nil
	}
	if actual == nil || expected == nil {
		return false, nil
	}

	// Try direct comparison first
	if actual == expected {
		return true, nil
	}

	// Try string comparison
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	if actualStr == expectedStr {
		return true, nil
	}

	// Try numeric comparison if both can be converted to numbers
	if actualFloat, err1 := e.convertToFloat(actual); err1 == nil {
		if expectedFloat, err2 := e.convertToFloat(expected); err2 == nil {
			return actualFloat == expectedFloat, nil
		}
	}

	return false, nil
}

// containsString checks if actual contains expected as a substring
func (e *BasicEvaluator) containsString(actual, expected interface{}) (bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)

	return strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedStr)), nil
}

// matchesRegex checks if actual matches the expected regex pattern
func (e *BasicEvaluator) matchesRegex(actual, expected interface{}) (bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	patternStr := fmt.Sprintf("%v", expected)

	regex, err := regexp.Compile(patternStr)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern %s: %w", patternStr, err)
	}

	return regex.MatchString(actualStr), nil
}

// inArray checks if actual value is in the expected array
func (e *BasicEvaluator) inArray(actual, expected interface{}) (bool, error) {
	// Expected should be an array/slice
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.Kind() != reflect.Slice && expectedValue.Kind() != reflect.Array {
		return false, fmt.Errorf("expected value must be an array or slice for 'in' operator")
	}

	actualStr := fmt.Sprintf("%v", actual)

	for i := 0; i < expectedValue.Len(); i++ {
		elemStr := fmt.Sprintf("%v", expectedValue.Index(i).Interface())
		if actualStr == elemStr {
			return true, nil
		}
	}

	return false, nil
}

// applyTimeWindow applies time-based aggregation (simplified implementation)
func (e *BasicEvaluator) applyTimeWindow(field, timeWindow string, data map[string]interface{}) (interface{}, error) {
	// Parse time window
	duration, err := time.ParseDuration(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("invalid time window format: %w", err)
	}

	// For now, just return the current value
	// In a real implementation, this would aggregate data over the time window
	// This would require historical data storage and querying
	e.logger.GetSlogger().Debug("Time window evaluation not fully implemented",
		slog.String("field", field),
		slog.String("window", timeWindow),
		slog.String("duration", duration.String()))

	return data[field], nil
}
