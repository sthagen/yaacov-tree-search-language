// Copyright 2019 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: 2019 Nimrod Shneor <nimrodshn@gmail.com>
// Author: 2019 Yaacov Zamir <kobi.zamir@gmail.com>

// Package semantics implements TSL tree semantics.
package semantics

import (
	"fmt"

	"github.com/yaacov/tree-search-language/v6/pkg/tsl"
)

// EvalFunc is a key evaluation function type
type EvalFunc = func(string) (interface{}, bool)

// Walk traverses the TSL tree and implements search semantics.
//
// Users can call the Walk method to check if a document compiles to `true` or `false`
// when applied to a tsl tree.
//
// Example:
//
//	record := map[string]interface{} {
//		"title":       "A good book",
//		"author":      "Joe",
//		"spec.pages":  14,
//		"spec.rating": 5,
//		"created_at":  time.Now(),
//		"is_active":   true,
//	}
//
//	// evalFactory creates an evaluation function for a data record.
//	func evalFactory(r map[string]interface{}) semantics.EvalFunc {
//		// Returns:
//		// A function (semantics.EvalFunc) that gets a `key` for a record and returns
//		// the value of the document for that key.
//		// If no value can be found for this `key` in our record, it will return
//		// ok = false, if value is found it will return ok = true.
//		return func(k string) (interface{}, bool) {
//			if v, ok := book[k]; ok {
//				return v, true
//			}
//			return nil, false
//		}
//	}
//
//	// Check if our record complies with our tsl tree.
//	//
//	// For example:
//	//   if our tsl tree represents the phrase "author = 'Joe' and created_at > '2023-01-01'"
//	//   we will get the boolean value `true` for our record.
//	//
//	//   if our tsl tree represents the phrase "spec.pages > 50"
//	//   we will get the boolean value `false` for our record.
//	eval := evalFactory(record)
//	compliance, err = semantics.Walk(tree, eval)
func Walk(n *tsl.TSLNode, eval EvalFunc) (interface{}, error) {
	if n == nil {
		return nil, nil
	}

	switch n.Type() {
	case tsl.KindIdentifier:
		return handleIdentifier(n, eval)
	case tsl.KindBinaryExpr:
		return handleBinaryExpression(n, eval)
	case tsl.KindUnaryExpr:
		return handleUnaryExpression(n, eval)
	case tsl.KindArrayLiteral:
		return handleArrayLiteral(n, eval)
	case tsl.KindNullLiteral:
		// null literal should be handled by the is expression
		return nil, nil
	default:
		return n.Value(), nil
	}
}

func handleBinaryExpression(n *tsl.TSLNode, eval EvalFunc) (interface{}, error) {
	exprOp, ok := n.Value().(tsl.TSLExpressionOp)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "TSLExpressionOp", Got: fmt.Sprintf("%T", n.Value())}
	}

	// lets walk the right side of the expression
	rightVal, err := Walk(exprOp.Right, eval)
	if err != nil {
		return nil, err
	}

	// lets walk the left side of the expression
	leftVal, err := Walk(exprOp.Left, eval)
	if err != nil {
		return nil, err
	}

	// handle the operator
	switch exprOp.Operator {
	case tsl.OpEQ:
		return leftVal == rightVal, nil
	case tsl.OpNE:
		return leftVal != rightVal, nil
	case tsl.OpLT, tsl.OpLE, tsl.OpGT, tsl.OpGE:
		return evaluateCompareExpressions(exprOp.Operator, leftVal, rightVal)
	case tsl.OpREQ:
		return evaluateRegexMatch(leftVal, rightVal)
	case tsl.OpRNE:
		matched, err := evaluateRegexMatch(leftVal, rightVal)
		if err != nil {
			return nil, err
		}
		return !matched, nil
	case tsl.OpAnd, tsl.OpOr:
		return evaluateLogicalExpression(exprOp.Operator, leftVal, rightVal)
	case tsl.OpLike:
		return evaluateLikePattern(leftVal, rightVal)
	case tsl.OpILike:
		return evaluateIlikePattern(leftVal, rightVal)
	case tsl.OpIn:
		// Try to extract the array values from the right side of the expression
		rightArray, ok := rightVal.([]interface{})
		if !ok {
			return nil, tsl.TypeMismatchError{Expected: "array", Got: fmt.Sprintf("%T", rightVal)}
		}

		return isValueInArray(leftVal, rightArray)
	case tsl.OpBetween:
		// Try to extract the array values from the right side of the expression
		rightArray, ok := rightVal.([]interface{})
		if !ok {
			return nil, tsl.TypeMismatchError{Expected: "array", Got: fmt.Sprintf("%T", rightVal)}
		}
		if len(rightArray) != 2 {
			return nil, tsl.TypeMismatchError{Expected: "min and max values", Got: fmt.Sprintf("%d values", len(rightArray))}
		}

		return isValueInRange(leftVal, rightArray[0], rightArray[1])
	case tsl.OpIs: // is null
		if rightVal == nil {
			return leftVal == nil, nil
		}
		return false, nil
	case tsl.OpPlus, tsl.OpMinus, tsl.OpStar, tsl.OpSlash, tsl.OpPercent:
		return evaluateMathExpression(exprOp.Operator, leftVal, rightVal)
	default:
		return nil, tsl.UnexpectedOperatorError{Operator: exprOp.Operator}
	}
}

func evaluateMathExpression(operator tsl.Operator, leftVal, rightVal interface{}) (interface{}, error) {
	leftNum, ok := toFloat64(leftVal)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
	}
	rightNum, ok := toFloat64(rightVal)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
	}

	switch operator {
	case tsl.OpPlus:
		return leftNum + rightNum, nil
	case tsl.OpMinus:
		return leftNum - rightNum, nil
	case tsl.OpStar:
		return leftNum * rightNum, nil
	case tsl.OpSlash:
		if rightNum == 0 {
			return nil, tsl.DivisionByZeroError{Operation: "division"}
		}
		return leftNum / rightNum, nil
	case tsl.OpPercent:
		if rightNum == 0 {
			return nil, tsl.DivisionByZeroError{Operation: "modulus"}
		}
		return float64(int64(leftNum) % int64(rightNum)), nil
	default:
		return nil, tsl.UnexpectedOperatorError{Operator: operator}
	}
}

func evaluateLogicalExpression(operator tsl.Operator, leftVal, rightVal interface{}) (interface{}, error) {
	leftBool, ok := leftVal.(bool)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", leftVal)}
	}
	rightBool, ok := rightVal.(bool)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", rightVal)}
	}

	switch operator {
	case tsl.OpAnd:
		return leftBool && rightBool, nil
	case tsl.OpOr:
		return leftBool || rightBool, nil
	default:
		return nil, tsl.UnexpectedOperatorError{Operator: operator}
	}
}

func evaluateCompareExpressions(operator tsl.Operator, leftVal, rightVal interface{}) (interface{}, error) {
	leftNum, leftIsNum := toFloat64(leftVal)
	rightNum, rightIsNum := toFloat64(rightVal)
	leftDate, leftIsDate := toDate(leftVal)
	rightDate, rightIsDate := toDate(rightVal)

	if leftIsNum && rightIsNum {
		switch operator {
		case tsl.OpLT:
			return leftNum < rightNum, nil
		case tsl.OpLE:
			return leftNum <= rightNum, nil
		case tsl.OpGT:
			return leftNum > rightNum, nil
		case tsl.OpGE:
			return leftNum >= rightNum, nil
		}
	} else if leftIsDate && rightIsDate {
		switch operator {
		case tsl.OpLT:
			return leftDate.Before(rightDate), nil
		case tsl.OpLE:
			return leftDate.Before(rightDate) || leftDate.Equal(rightDate), nil
		case tsl.OpGT:
			return leftDate.After(rightDate), nil
		case tsl.OpGE:
			return leftDate.After(rightDate) || leftDate.Equal(rightDate), nil
		}
	} else {
		return nil, tsl.TypeMismatchError{Expected: "number or date", Got: fmt.Sprintf("%T and %T", leftVal, rightVal)}
	}
	return nil, nil
}

func handleUnaryExpression(n *tsl.TSLNode, eval EvalFunc) (interface{}, error) {
	exprOp, ok := n.Value().(tsl.TSLExpressionOp)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "TSLExpressionOp", Got: fmt.Sprintf("%T", n.Value())}
	}

	// lets walk the right side of the expression
	rightVal, err := Walk(exprOp.Right, eval)
	if err != nil {
		return nil, err
	}

	switch exprOp.Operator {
	case tsl.OpNot:
		rightBool, ok := rightVal.(bool)
		if !ok {
			return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", rightVal)}
		}
		return !rightBool, nil
	case tsl.OpMinus:
		rightNum, ok := toFloat64(rightVal)
		if !ok {
			return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
		}
		return -rightNum, nil
	default:
		return nil, tsl.UnexpectedOperatorError{Operator: exprOp.Operator}
	}
}

func handleArrayLiteral(n *tsl.TSLNode, eval EvalFunc) (interface{}, error) {
	exprOp, ok := n.Value().(tsl.TSLArrayLiteral)
	if !ok {
		return nil, tsl.TypeMismatchError{Expected: "TSLArrayLiteral", Got: fmt.Sprintf("%T", n.Value())}
	}
	values := make([]interface{}, len(exprOp.Values))
	for i, v := range exprOp.Values {
		val, err := Walk(v, eval)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return values, nil
}
