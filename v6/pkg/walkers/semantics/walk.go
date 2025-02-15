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
		return EvalIdent(n, eval)
	case tsl.KindBinaryExpr:
		exprOp := n.Value().(tsl.TSLExpressionOp)

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
		case tsl.OpLT:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum < rightNum, nil
		case tsl.OpLE:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum <= rightNum, nil
		case tsl.OpGT:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum > rightNum, nil
		case tsl.OpGE:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum >= rightNum, nil
		case tsl.OpREQ:
			return EvalRegexp(leftVal, rightVal)
		case tsl.OpRNE:
			matched, err := EvalRegexp(leftVal, rightVal)
			if err != nil {
				return nil, err
			}
			return !matched, nil
		case tsl.OpAnd:
			leftBool, ok := leftVal.(bool)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightBool, ok := rightVal.(bool)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftBool && rightBool, nil
		case tsl.OpOr:
			leftBool, ok := leftVal.(bool)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightBool, ok := rightVal.(bool)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "boolean", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftBool || rightBool, nil
		case tsl.OpLike:
			return EvalLike(leftVal, rightVal)
		case tsl.OpILike:
			return EvalILike(leftVal, rightVal)
		case tsl.OpIn:
			// Try to extract the array values from the right side of the expression
			rightArray, ok := rightVal.([]interface{})
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "array", Got: fmt.Sprintf("%T", rightVal)}
			}

			return EvalIn(leftVal, rightArray)
		case tsl.OpBetween:
			// Try to extract the array values from the right side of the expression
			rightArray, ok := rightVal.([]interface{})
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "array", Got: fmt.Sprintf("%T", rightVal)}
			}
			if len(rightArray) != 2 {
				return nil, tsl.TypeMismatchError{Expected: "2 values", Got: fmt.Sprintf("%d values", len(rightArray))}
			}

			return EvalBetween(leftVal, rightArray[0], rightArray[1])
		case tsl.OpIs:
			if rightVal == nil {
				return leftVal == nil, nil
			}
			return false, nil
		case tsl.OpPlus:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum + rightNum, nil
		case tsl.OpMinus:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}

			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum - rightNum, nil
		case tsl.OpStar:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			return leftNum * rightNum, nil
		case tsl.OpSlash:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {

				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			if rightNum == 0 {
				return nil, tsl.DivisionByZeroError{Operation: "division"}
			}
			return leftNum / rightNum, nil
		case tsl.OpPercent:
			leftNum, ok := toFloat64(leftVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", leftVal)}
			}
			rightNum, ok := toFloat64(rightVal)
			if !ok {
				return nil, tsl.TypeMismatchError{Expected: "number", Got: fmt.Sprintf("%T", rightVal)}
			}
			if rightNum == 0 {
				return nil, tsl.DivisionByZeroError{Operation: "modulus"}
			}
			return float64(int64(leftNum) % int64(rightNum)), nil
		default:
			return nil, tsl.UnexpectedOperatorError{Operator: exprOp.Operator}
		}

	case tsl.KindUnaryExpr:
		exprOp := n.Value().(tsl.TSLExpressionOp)

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

	case tsl.KindArrayLiteral:
		// array literals should be handled by the binary expression
		return nil, nil

	case tsl.KindNullLiteral:
		// null literal should be handled by the is expression
		return nil, nil

	default:
		return n.Value(), nil
	}
}
