package node

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// CredentialResolver is a global hook that can be set by the credential module
// to allow ParseTemplate to resolve sensitive values from the database.
var CredentialResolver func(string) (string, error)

func ParseTemplate(template string, variables map[string]interface{}) string {
	// Step 1: Evaluate arithmetic expressions FIRST: {{var + 1}}, {{var * 2}}, etc.
	// This must be done BEFORE variable replacement so we can detect {{varName + operator}}
	template = EvaluateArithmetic(template, variables)

	// Step 2: Replace standard variables (format: {{var}} or {{ var }})
	for k, v := range variables {
		// Create regex that allows optional spaces: {{\s*key\s*}}
		re := regexp.MustCompile(fmt.Sprintf(`\{\{\s*%s\s*\}\}`, regexp.QuoteMeta(k)))
		
		// Convert value to string - use JSON for complex types
		var strVal string
		switch val := v.(type) {
		case string:
			strVal = val
		case int, int64, float64, bool:
			strVal = fmt.Sprintf("%v", val)
		case nil:
			strVal = ""
		default:
			// For maps, slices, and other complex types, use JSON
			if jsonBytes, err := json.Marshal(val); err == nil {
				strVal = string(jsonBytes)
			} else {
				strVal = fmt.Sprintf("%v", val)
			}
		}
		
		template = re.ReplaceAllString(template, strVal)
	}

	// Step 3: Replace environment variables: {{env.KEY}} or {{ env.KEY }}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			k, v := pair[0], pair[1]
			re := regexp.MustCompile(fmt.Sprintf(`\{\{\s*env\.%s\s*\}\}`, regexp.QuoteMeta(k)))
			template = re.ReplaceAllString(template, v)
		}
	}

	// Step 4: Resolve Credentials using the global resolver (if set)
	if CredentialResolver != nil {
		// Use regex to find {{ cred.KEY }} patterns
		re := regexp.MustCompile(`\{\{\s*(cred\..+?)\s*\}\}`)
		template = re.ReplaceAllStringFunc(template, func(match string) string {
			matches := re.FindStringSubmatch(match)
			if len(matches) < 2 {
				return match
			}
			key := strings.TrimSpace(matches[1])

			// Try to resolve
			val, err := CredentialResolver(key)
			if err != nil {
				// If resolution fails, keep the placeholder
				return match
			}
			return val
		})
	}


	return template
}

// EvaluateArithmetic evaluates arithmetic expressions in the format {{var + 1}}, {{var * 2}}, etc.
func EvaluateArithmetic(template string, variables map[string]interface{}) string {
	// Regex to match {{expression}} where expression contains arithmetic operators
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*([\+\-\*\/])\s*([a-zA-Z0-9_\.]+)\s*\}\}`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match
		}

		varName := matches[1]
		operator := matches[2]
		operand2Str := matches[3]

		// Get variable value
		varVal, exists := variables[varName]
		if !exists {
			return match // Variable not found, return as is
		}

		// Convert varVal to float64
		var1, err := toNumber(varVal)
		if err != nil {
			return match // Cannot convert to number, return as is
		}

		// Get operand2 value (could be a number or variable)
		var var2 float64
		if operand2Val, exists := variables[operand2Str]; exists {
			var2, err = toNumber(operand2Val)
			if err != nil {
				return match
			}
		} else {
			// Try to parse as literal number
			var2, err = strconv.ParseFloat(operand2Str, 64)
			if err != nil {
				return match
			}
		}

		// Perform arithmetic operation
		var result float64
		switch operator {
		case "+":
			result = var1 + var2
		case "-":
			result = var1 - var2
		case "*":
			result = var1 * var2
		case "/":
			if var2 == 0 {
				return match // Division by zero
			}
			result = var1 / var2
		}

		// Return result as integer if it's a whole number
		if result == float64(int64(result)) {
			return fmt.Sprintf("%d", int64(result))
		}
		return fmt.Sprintf("%g", result)
	})
}

// toNumber converts an interface{} to float64
func toNumber(val interface{}) (float64, error) {
	switch v := val.(type) {
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
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", val)
	}
}

// FieldValue returns the Value of a field by key from a NodeAction.
// Falls back to Default if Value is nil.
func FieldValue(action NodeAction, key string) any {
	for _, f := range action.Fields {
		if f.Key == key {
			if f.Value != nil {
				return f.Value
			}
			return f.Default
		}
	}
	return nil
}

// ActionFromMap creates a NodeAction with fields populated from a map (for testing).
func ActionFromMap(m map[string]any) NodeAction {
	action := NodeAction{}
	
	// Extract action key if present
	if actionKey, ok := m["action"].(string); ok {
		action.Key = actionKey
	}
	
	// Convert map to fields
	fields := make([]NodeField, 0, len(m))
	for k, v := range m {
		fields = append(fields, NodeField{Key: k, Value: v})
	}
	action.Fields = fields
	
	return action
}