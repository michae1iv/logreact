package rule_manager

import (
	"correlator/db"
	"strconv"
	"strings"
)

// Function for obtaining value from the logs by key and checking if it contains expected value.
//
// # If value is type of int, float or bool checking if values are equal
//
// for operator :
func CheckValueFromLog(logData map[string]interface{}, token string, condValue string) bool {
	if value, exists := logData[token]; exists {
		if strValue, ok := value.(string); ok {
			return strings.Contains(strValue, condValue)
		} else if floatValue, ok := value.(float64); ok {
			if newVal, err := strconv.ParseFloat(condValue, 64); err == nil {
				return newVal == floatValue
			}
		} else if boolValue, ok := value.(bool); ok {
			strValue := strconv.FormatBool(boolValue)
			return strValue == condValue
		}
	}
	return false
}

// Compares value with field, value and field must be equal
//
// for operator == or =
func CheckPerciseValueFromLog(logData map[string]interface{}, token string, condValue string) bool {
	if value, exists := logData[token]; exists {
		if strValue, ok := value.(string); ok {
			return strValue == condValue
		} else if floatValue, ok := value.(float64); ok {
			if newVal, err := strconv.ParseFloat(condValue, 64); err == nil {
				return newVal == floatValue
			}
		} else if boolValue, ok := value.(bool); ok {
			strValue := strconv.FormatBool(boolValue)
			return strValue == condValue
		}
	}
	return false
}

// Function which checks if value in list
//
// for operator ->, !->  and contains
func InList(logData map[string]interface{}, token string, list string) bool {
	var items []string

	// if list in format: "[val1, val2]"
	if strings.HasPrefix(list, "[") && strings.HasSuffix(list, "]") {
		trimmed := strings.Trim(list, "[]")
		for _, item := range strings.Split(trimmed, ",") {
			items = append(items, strings.TrimSpace(item))
		}
	} else {
		lst := db.List{}
		if res := db.DB.Where(db.List{List_name: list}).First(&lst); res.RowsAffected < 1 {
			return false
		}
		items = lst.Phrases
	}

	// Checking insertion
	for _, item := range items {
		val, ok := logData[token].(string)
		if ok && item == val {
			return true
		}
	}
	return false
}

// Function for obtaining value from the logs by key, if key doesn't exist returns nothing
func getValueFromLog(logData map[string]interface{}, token string) interface{} {
	if value, exists := logData[token]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		} else if floatValue, ok := value.(float64); ok {
			return floatValue
		} else if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return nil
}

func isFieldExists(logData map[string]interface{}, token string) bool {
	if _, exists := logData[token]; exists {
		return true
	} else {
		return false
	}
}

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}
