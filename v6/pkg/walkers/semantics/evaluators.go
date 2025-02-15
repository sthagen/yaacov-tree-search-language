package semantics

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/yaacov/tree-search-language/v6/pkg/tsl"
)

// EvalIdent evaluates an identifier node using the provided eval function
func EvalIdent(n *tsl.TSLNode, eval EvalFunc) (interface{}, error) {
	if n == nil {
		return nil, nil
	}

	// If not an identifier, just return the node
	if n.Type() != tsl.KindIdentifier {
		return Walk(n, eval)
	}

	// Get identifier name
	identName, ok := n.Value().(string)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "identifier", Got: fmt.Sprintf("%T", n.Value())}
	}

	// Get value using eval function
	value, exists := eval(identName)
	if !exists {
		return nil, tsl.KeyNotFoundError{Key: identName}
	}

	// Check if value is a string and matches ISO date and time format
	if strValue, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, strValue); err == nil {
			return t, nil
		}
	}

	return value, nil
}

// EvalIn checks if a value is in a list of values
// Supports string, float64, time.Time, and bool values
func EvalIn(value interface{}, arr []interface{}) (bool, error) {
	if value == nil || arr == nil {
		return false, nil
	}

	switch value.(type) {
	case string, float64, time.Time, bool:
		// Continue with existing type-specific checks
		switch v := value.(type) {
		case string:
			for _, item := range arr {
				if str, ok := item.(string); ok && str == v {
					return true, nil
				}
			}
		case float64:
			for _, item := range arr {
				switch i := item.(type) {
				case float64:
					if i == v {
						return true, nil
					}
				case int:
					if float64(i) == v {
						return true, nil
					}
				}
			}
		case time.Time:
			for _, item := range arr {
				if t, ok := item.(time.Time); ok && t.Equal(v) {
					return true, nil
				}
			}
		case bool:
			for _, item := range arr {
				if b, ok := item.(bool); ok && b == v {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		return false, &tsl.TypeMismatchError{
			Expected: "string, float64, time.Time, or bool",
			Got:      value,
		}
	}
}

// EvalBetween checks if a value is within a range (inclusive)
// Supports both numeric values and time.Time comparisons
func EvalBetween(value, min, max interface{}) (bool, error) {
	switch v := value.(type) {
	case float64:
		minVal, okMin := toFloat64(min)
		maxVal, okMax := toFloat64(max)
		if okMin && okMax {
			return v >= minVal && v <= maxVal, nil
		}
		// Try int conversion if float64 conversion failed
		if !okMin {
			if minInt, ok := min.(int); ok {
				minVal = float64(minInt)
				okMin = true
			}
		}
		if !okMax {
			if maxInt, ok := max.(int); ok {
				maxVal = float64(maxInt)
				okMax = true
			}
		}
		if !okMin || !okMax {
			return false, &tsl.TypeMismatchError{
				Expected: "numeric values",
				Got:      value,
			}
		}
		return v >= minVal && v <= maxVal, nil
	case time.Time:
		minTime, okMin := min.(time.Time)
		maxTime, okMax := max.(time.Time)
		if !okMin || !okMax {
			return false, &tsl.TypeMismatchError{
				Expected: "time values",
				Got:      value,
			}
		}
		return !v.Before(minTime) && !v.After(maxTime), nil
	}
	return false, &tsl.TypeMismatchError{
		Expected: "numeric or time.Time values",
		Got:      value,
	}
}

// EvalLike performs pattern matching with SQL LIKE semantics
// Supports % for any sequence of characters and _ for single character
func EvalLike(value interface{}, pattern interface{}) (bool, error) {
	valueStr, okValue := value.(string)
	patternStr, okPattern := pattern.(string)

	if !okValue {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      value,
		}
	}
	if !okPattern {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      pattern,
		}
	}

	// Convert SQL LIKE pattern to regex pattern
	patternStr = strings.ReplaceAll(patternStr, "%", ".*")
	patternStr = strings.ReplaceAll(patternStr, "_", ".")
	matched, _ := regexp.MatchString("^"+patternStr+"$", valueStr)
	return matched, nil
}

// EvalILike performs case-insensitive pattern matching with SQL LIKE semantics
func EvalILike(value interface{}, pattern interface{}) (bool, error) {
	valueStr, okValue := value.(string)
	patternStr, okPattern := pattern.(string)

	if !okValue {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      value,
		}
	}
	if !okPattern {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      pattern,
		}
	}

	return EvalLike(strings.ToLower(valueStr), strings.ToLower(patternStr))
}

// EvalRegexp evaluates if a string matches a regular expression pattern
func EvalRegexp(value interface{}, pattern interface{}) (bool, error) {
	valueStr, okValue := value.(string)
	patternStr, okPattern := pattern.(string)

	if !okValue {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      value,
		}
	}
	if !okPattern {
		return false, &tsl.TypeMismatchError{
			Expected: "string",
			Got:      pattern,
		}
	}

	matched, err := regexp.MatchString(patternStr, valueStr)
	if err != nil {
		return false, err
	}
	return matched, nil
}
