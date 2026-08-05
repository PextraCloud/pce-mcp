/*
Copyright 2026 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package pce

import (
	"reflect"
	"testing"
)

func TestApplyStringFilter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		filter   string
		expected bool
	}{
		{
			name:     "empty filter returns true",
			value:    "test",
			filter:   "",
			expected: true,
		},
		{
			name:     "exact match (case insensitive)",
			value:    "Organization",
			filter:   "organization",
			expected: true,
		},
		{
			name:     "partial match (case insensitive)",
			value:    "Datacenter-1",
			filter:   "data",
			expected: true,
		},
		{
			name:     "no match",
			value:    "test",
			filter:   "xyz",
			expected: false,
		},
		{
			name:     "empty value with non-empty filter returns false",
			value:    "",
			filter:   "test",
			expected: false,
		},
		{
			name:     "both empty strings returns true",
			value:    "",
			filter:   "",
			expected: true,
		},
		{
			name:     "case insensitive partial match in middle",
			value:    "MyCluster-123",
			filter:   "cluster",
			expected: true,
		},
		{
			name:     "case insensitive partial match at end",
			value:    "Node-456",
			filter:   "456",
			expected: true,
		},
		{
			name:     "case insensitive partial match at start",
			value:    "Server-789",
			filter:   "server",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyStringFilter(tt.value, tt.filter)
			if result != tt.expected {
				t.Errorf("applyStringFilter(%q, %q) = %v; expected %v",
					tt.value, tt.filter, result, tt.expected)
			}
		})
	}
}

func TestConvertFilterToInt(t *testing.T) {
	tests := []struct {
		name     string
		filter   map[string]any
		expected map[string]int
	}{
		{
			name:     "nil filter returns nil",
			filter:   nil,
			expected: nil,
		},
		{
			name:     "empty map returns empty map",
			filter:   map[string]any{},
			expected: map[string]int{},
		},
		{
			name: "valid float64 values converted to int",
			filter: map[string]any{
				"gt": 5.0,
				"lt": 10.0,
			},
			expected: map[string]int{
				"gt": 5,
				"lt": 10,
			},
		},
		{
			name: "only float64 converted",
			filter: map[string]any{
				"gt":     5.0,
				"lt":     "string",
				"eq":     3.0,
				"string": "value",
				"bool":   true,
				"nil":    nil,
			},
			expected: map[string]int{
				"gt": 5,
				"eq": 3,
			},
		},
		{
			name: "zero value converted correctly",
			filter: map[string]any{
				"eq": 0.0,
			},
			expected: map[string]int{
				"eq": 0,
			},
		},
		{
			name: "negative numbers converted correctly",
			filter: map[string]any{
				"lt": -5.0,
			},
			expected: map[string]int{
				"lt": -5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFilterToInt(tt.filter)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("convertFilterToInt(%v) = %v; expected %v", tt.filter, result, tt.expected)
			}
		})
	}
}

func TestApplyNumberFilter(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		filter   map[string]int
		expected bool
	}{
		{
			name:     "nil filter returns true",
			value:    5,
			filter:   nil,
			expected: true,
		},
		{
			name:     "empty filter returns true",
			value:    5,
			filter:   map[string]int{},
			expected: true,
		},
		{
			name:  "gt, value greater than filter",
			value: 10,
			filter: map[string]int{
				"gt": 5,
			},
			expected: true,
		},
		{
			name:  "gt, value equal to filter",
			value: 5,
			filter: map[string]int{
				"gt": 5,
			},
			expected: false,
		},
		{
			name:  "gt, value less than filter",
			value: 3,
			filter: map[string]int{
				"gt": 5,
			},
			expected: false,
		},
		{
			name:  "lt, value less than filter",
			value: 3,
			filter: map[string]int{
				"lt": 5,
			},
			expected: true,
		},
		{
			name:  "lt, value equal to filter",
			value: 5,
			filter: map[string]int{
				"lt": 5,
			},
			expected: false,
		},
		{
			name:  "lt, value greater than filter",
			value: 10,
			filter: map[string]int{
				"lt": 5,
			},
			expected: false,
		},
		{
			name:  "eq, value equals filter",
			value: 5,
			filter: map[string]int{
				"eq": 5,
			},
			expected: true,
		},
		{
			name:  "eq, value does not equal filter",
			value: 3,
			filter: map[string]int{
				"eq": 5,
			},
			expected: false,
		},
		{
			name:  "gt and lt, value within range",
			value: 5,
			filter: map[string]int{
				"gt": 3,
				"lt": 10,
			},
			expected: true,
		},
		{
			name:  "gt and lt, value below range",
			value: 3,
			filter: map[string]int{
				"gt": 3,
				"lt": 10,
			},
			expected: false,
		},
		{
			name:  "gt and lt, value above range",
			value: 10,
			filter: map[string]int{
				"gt": 3,
				"lt": 10,
			},
			expected: false,
		},
		{
			name:  "gt and lt, value equals eq",
			value: 5,
			filter: map[string]int{
				"gt": 3,
				"lt": 7,
				"eq": 5,
			},
			expected: true,
		},
		{
			name:  "gt, negative value greater than filter",
			value: -3,
			filter: map[string]int{
				"gt": -5,
			},
			expected: true,
		},
		{
			name:  "eq, zero value",
			value: 0,
			filter: map[string]int{
				"eq": 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyNumberFilter(tt.value, tt.filter)
			if result != tt.expected {
				t.Errorf("applyNumberFilter(%d, %v) = %v; expected %v",
					tt.value, tt.filter, result, tt.expected)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestApplyBoolFilter(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		filter   *bool
		expected bool
	}{
		{
			name:     "nil filter returns true",
			value:    true,
			filter:   nil,
			expected: true,
		},
		{
			name:     "value true with filter true returns true",
			value:    true,
			filter:   boolPtr(true),
			expected: true,
		},
		{
			name:     "value false with filter true returns false",
			value:    false,
			filter:   boolPtr(true),
			expected: false,
		},
		{
			name:     "value true with filter false returns false",
			value:    true,
			filter:   boolPtr(false),
			expected: false,
		},
		{
			name:     "value false with filter false returns true",
			value:    false,
			filter:   boolPtr(false),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyBoolFilter(tt.value, tt.filter)
			if result != tt.expected {
				t.Errorf("applyBoolFilter(%v, %v) = %v; expected %v",
					tt.value, tt.filter, result, tt.expected)
			}
		})
	}
}
