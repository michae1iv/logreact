package rule_manager

import (
	"correlator/regulars"
	"fmt"
	"strings"
)

// Priority of operations
var Preference = map[string]int{
	"->":       2, // If value contains in the list
	"!->":      2, // If value doesn't contains in the list
	"!=":       2,
	"==":       2,
	"=":        2,
	":":        2,
	"!:":       2,
	"CONTAINS": 2,
	"contains": 2,
	"AND":      1,
	"and":      1,
	"OR":       0,
	"or":       0,
}

// Checking if token is operator
func isOperator(token string) bool {
	_, exists := Preference[token]
	return exists
}

// CHeckin if token is operand
func isOperand(token string) bool {
	return !isOperator(token) && !strings.HasPrefix(token, "(") && !strings.HasSuffix(token, ")")
}

// Function to convert condition into reverse polish notation
func ParseCondToRPN(condition string) ([]string, error) {
	output := []string{}
	stack := []string{}
	operands := 0
	operators := 0

	tokens := regulars.Token_reg.FindAllString(condition, -1)

	for _, token := range tokens {
		if token == "" {
			continue
		}
		if value := regulars.Value_reg.FindStringSubmatch(token); len(value) == 2 {
			token = value[1]
		}
		if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			// if token - closed bracket, push out of stack all operators until open bracket
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, fmt.Errorf("brackets mismatch")
			}
			stack = stack[:len(stack)-1] // Deleting open bracket
		} else if isOperator(token) {
			// if token is operator
			operators++
			for len(stack) > 0 && isOperator(stack[len(stack)-1]) && Preference[stack[len(stack)-1]] >= Preference[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			operands++
			output = append(output, token) // Adding opperands
		}
	}

	if operands-1 > operators { // Checking if amount of operands and operators are valid
		return nil, fmt.Errorf("there are extra operators")
	} else if operands-1 < operators {
		return nil, fmt.Errorf("there are extra operands")
	}

	// Adding remaining operators from stack in output array
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, fmt.Errorf("brackets mismatch")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return output, nil
}
