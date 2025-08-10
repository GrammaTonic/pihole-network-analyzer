package alerts

import (
	"testing"

	"pihole-analyzer/internal/logger"
)

// TestBasicEvaluatorComparisons tests basic comparison operations
func TestBasicEvaluatorComparisons(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name      string
		condition AlertCondition
		data      map[string]interface{}
		expected  bool
		expectErr bool
	}{
		{
			name: "greater than - true",
			condition: AlertCondition{
				Field:    "value",
				Operator: "gt",
				Value:    100,
			},
			data:     map[string]interface{}{"value": 150},
			expected: true,
		},
		{
			name: "greater than - false",
			condition: AlertCondition{
				Field:    "value",
				Operator: "gt",
				Value:    100,
			},
			data:     map[string]interface{}{"value": 50},
			expected: false,
		},
		{
			name: "less than - true",
			condition: AlertCondition{
				Field:    "value",
				Operator: "lt",
				Value:    100,
			},
			data:     map[string]interface{}{"value": 50},
			expected: true,
		},
		{
			name: "equal - true",
			condition: AlertCondition{
				Field:    "status",
				Operator: "eq",
				Value:    "active",
			},
			data:     map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "equal - false",
			condition: AlertCondition{
				Field:    "status",
				Operator: "eq",
				Value:    "active",
			},
			data:     map[string]interface{}{"status": "inactive"},
			expected: false,
		},
		{
			name: "not equal - true",
			condition: AlertCondition{
				Field:    "status",
				Operator: "ne",
				Value:    "active",
			},
			data:     map[string]interface{}{"status": "inactive"},
			expected: true,
		},
		{
			name: "contains - true",
			condition: AlertCondition{
				Field:    "message",
				Operator: "contains",
				Value:    "error",
			},
			data:     map[string]interface{}{"message": "connection error occurred"},
			expected: true,
		},
		{
			name: "contains - false",
			condition: AlertCondition{
				Field:    "message",
				Operator: "contains",
				Value:    "error",
			},
			data:     map[string]interface{}{"message": "success"},
			expected: false,
		},
		{
			name: "field not found",
			condition: AlertCondition{
				Field:    "missing_field",
				Operator: "gt",
				Value:    100,
			},
			data:     map[string]interface{}{"value": 150},
			expected: false,
		},
		{
			name: "unsupported operator",
			condition: AlertCondition{
				Field:    "value",
				Operator: "unknown",
				Value:    100,
			},
			data:      map[string]interface{}{"value": 150},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateCondition(tt.condition, tt.data)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluatorTypeConversion tests type conversion in evaluator
func TestEvaluatorTypeConversion(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name      string
		condition AlertCondition
		data      map[string]interface{}
		expected  bool
	}{
		{
			name: "int to float comparison",
			condition: AlertCondition{
				Field:    "value",
				Operator: "gt",
				Value:    100.5,
			},
			data:     map[string]interface{}{"value": 101},
			expected: true,
		},
		{
			name: "string to int comparison",
			condition: AlertCondition{
				Field:    "value",
				Operator: "eq",
				Value:    "123",
			},
			data:     map[string]interface{}{"value": 123},
			expected: true,
		},
		{
			name: "float to int comparison",
			condition: AlertCondition{
				Field:    "value",
				Operator: "eq",
				Value:    100,
			},
			data:     map[string]interface{}{"value": 100.0},
			expected: true,
		},
		{
			name: "mixed type string comparison",
			condition: AlertCondition{
				Field:    "count",
				Operator: "eq",
				Value:    42,
			},
			data:     map[string]interface{}{"count": "42"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateCondition(tt.condition, tt.data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluatorMultipleConditions tests evaluating multiple conditions
func TestEvaluatorMultipleConditions(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name       string
		conditions []AlertCondition
		data       map[string]interface{}
		expected   bool
	}{
		{
			name: "all conditions true",
			conditions: []AlertCondition{
				{
					Field:    "count",
					Operator: "gt",
					Value:    10,
				},
				{
					Field:    "status",
					Operator: "eq",
					Value:    "active",
				},
			},
			data: map[string]interface{}{
				"count":  20,
				"status": "active",
			},
			expected: true,
		},
		{
			name: "one condition false",
			conditions: []AlertCondition{
				{
					Field:    "count",
					Operator: "gt",
					Value:    10,
				},
				{
					Field:    "status",
					Operator: "eq",
					Value:    "active",
				},
			},
			data: map[string]interface{}{
				"count":  5, // This fails the first condition
				"status": "active",
			},
			expected: false,
		},
		{
			name: "all conditions false",
			conditions: []AlertCondition{
				{
					Field:    "count",
					Operator: "gt",
					Value:    10,
				},
				{
					Field:    "status",
					Operator: "eq",
					Value:    "active",
				},
			},
			data: map[string]interface{}{
				"count":  5,          // Fails first condition
				"status": "inactive", // Fails second condition
			},
			expected: false,
		},
		{
			name:       "no conditions",
			conditions: []AlertCondition{},
			data:       map[string]interface{}{"count": 20},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateConditions(tt.conditions, tt.data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluatorRegex tests regex pattern matching
func TestEvaluatorRegex(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name      string
		condition AlertCondition
		data      map[string]interface{}
		expected  bool
		expectErr bool
	}{
		{
			name: "valid regex - match",
			condition: AlertCondition{
				Field:    "domain",
				Operator: "regex",
				Value:    `^.*\.malware\.com$`,
			},
			data:     map[string]interface{}{"domain": "suspicious.malware.com"},
			expected: true,
		},
		{
			name: "valid regex - no match",
			condition: AlertCondition{
				Field:    "domain",
				Operator: "regex",
				Value:    `^.*\.malware\.com$`,
			},
			data:     map[string]interface{}{"domain": "google.com"},
			expected: false,
		},
		{
			name: "invalid regex pattern",
			condition: AlertCondition{
				Field:    "domain",
				Operator: "regex",
				Value:    `[invalid`,
			},
			data:      map[string]interface{}{"domain": "test.com"},
			expected:  false,
			expectErr: true,
		},
		{
			name: "email pattern",
			condition: AlertCondition{
				Field:    "email",
				Operator: "regex",
				Value:    `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			data:     map[string]interface{}{"email": "user@example.com"},
			expected: true,
		},
		{
			name: "IP address pattern",
			condition: AlertCondition{
				Field:    "ip",
				Operator: "regex",
				Value:    `^192\.168\.`,
			},
			data:     map[string]interface{}{"ip": "192.168.1.100"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateCondition(tt.condition, tt.data)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluatorInArray tests array membership operations
func TestEvaluatorInArray(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name      string
		condition AlertCondition
		data      map[string]interface{}
		expected  bool
		expectErr bool
	}{
		{
			name: "in array - found",
			condition: AlertCondition{
				Field:    "status",
				Operator: "in",
				Value:    []string{"active", "pending", "warning"},
			},
			data:     map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "in array - not found",
			condition: AlertCondition{
				Field:    "status",
				Operator: "in",
				Value:    []string{"active", "pending", "warning"},
			},
			data:     map[string]interface{}{"status": "inactive"},
			expected: false,
		},
		{
			name: "not in array - true",
			condition: AlertCondition{
				Field:    "status",
				Operator: "not_in",
				Value:    []string{"active", "pending", "warning"},
			},
			data:     map[string]interface{}{"status": "inactive"},
			expected: true,
		},
		{
			name: "not in array - false",
			condition: AlertCondition{
				Field:    "status",
				Operator: "not_in",
				Value:    []string{"active", "pending", "warning"},
			},
			data:     map[string]interface{}{"status": "active"},
			expected: false,
		},
		{
			name: "in array - numbers",
			condition: AlertCondition{
				Field:    "code",
				Operator: "in",
				Value:    []int{200, 201, 202},
			},
			data:     map[string]interface{}{"code": 201},
			expected: true,
		},
		{
			name: "in array - invalid type",
			condition: AlertCondition{
				Field:    "status",
				Operator: "in",
				Value:    "not_an_array",
			},
			data:      map[string]interface{}{"status": "active"},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateCondition(tt.condition, tt.data)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluatorEdgeCases tests edge cases and error conditions
func TestEvaluatorEdgeCases(t *testing.T) {
	logger := logger.New(logger.DefaultConfig())
	evaluator := NewEvaluator(logger)

	tests := []struct {
		name      string
		condition AlertCondition
		data      map[string]interface{}
		expected  bool
		expectErr bool
	}{
		{
			name: "nil values comparison",
			condition: AlertCondition{
				Field:    "nullable_field",
				Operator: "eq",
				Value:    nil,
			},
			data:     map[string]interface{}{"nullable_field": nil},
			expected: true,
		},
		{
			name: "nil vs non-nil",
			condition: AlertCondition{
				Field:    "nullable_field",
				Operator: "eq",
				Value:    nil,
			},
			data:     map[string]interface{}{"nullable_field": "value"},
			expected: false,
		},
		{
			name: "boolean comparison",
			condition: AlertCondition{
				Field:    "enabled",
				Operator: "eq",
				Value:    true,
			},
			data:     map[string]interface{}{"enabled": true},
			expected: true,
		},
		{
			name: "boolean vs string",
			condition: AlertCondition{
				Field:    "enabled",
				Operator: "eq",
				Value:    "true",
			},
			data:     map[string]interface{}{"enabled": true},
			expected: true,
		},
		{
			name: "empty string contains",
			condition: AlertCondition{
				Field:    "message",
				Operator: "contains",
				Value:    "",
			},
			data:     map[string]interface{}{"message": "any message"},
			expected: true, // Empty string is contained in any string
		},
		{
			name: "case insensitive contains",
			condition: AlertCondition{
				Field:    "message",
				Operator: "contains",
				Value:    "ERROR",
			},
			data:     map[string]interface{}{"message": "connection error occurred"},
			expected: true, // Should be case insensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateCondition(tt.condition, tt.data)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
