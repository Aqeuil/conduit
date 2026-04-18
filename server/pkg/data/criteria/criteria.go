package criteria

import (
	"fmt"
	"strings"
)

// Criteria 查询标准
type Criteria interface {
	GetQuery() string
	GetValues() []any
}

func And(criteriaCollection ...Criteria) Criteria {
	return joinCriteria{
		criteriaCollection: criteriaCollection,
		separator:          "AND",
	}
}

func Or(criteriaCollection ...Criteria) Criteria {
	return joinCriteria{
		criteriaCollection: criteriaCollection,
		separator:          "OR",
	}
}

func Not(Criteria Criteria) Criteria {
	return notCriteria{
		Criteria,
	}
}

func IsNull(field string) Criteria {
	return stringCriteria(fmt.Sprintf("%s IS NULL", field))
}

func IsNotNull(field string) Criteria {
	return stringCriteria(fmt.Sprintf("%s IS NOT NULL", field))
}

func Binary[T any](field string, operator string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: operator,
		value:    value,
	}
}

func Equal[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "=",
		value:    value,
	}
}

// Eq alias for Equal
func Eq[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "=",
		value:    value,
	}
}

func GreaterThan[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: ">",
		value:    value,
	}
}

// Gt alias for GreaterThan
func Gt[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: ">",
		value:    value,
	}
}

func GreaterOrEqual[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: ">=",
		value:    value,
	}
}

// Gte alias for GreaterOrEqual
func Gte[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: ">=",
		value:    value,
	}
}

func LessThan[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "<",
		value:    value,
	}
}

// Lt alias for LessThan
func Lt[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "<",
		value:    value,
	}
}

func LessOrEqual[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "<=",
		value:    value,
	}
}

// Le alias for LessOrEqual
func Le[T comparable](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "<=",
		value:    value,
	}
}

func LeftLike[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[string]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "LIKE",
		value:    fmt.Sprintf("%v%%", value),
	}
}

// Llk alias for LeftLike
func Llk[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[string]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "LIKE",
		value:    fmt.Sprintf("%v%%", value),
	}
}

func Like[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[string]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "LIKE",
		value:    fmt.Sprintf("%%%v%%", value),
	}
}

// Lk alias for Like
func Lk[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[string]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "LIKE",
		value:    fmt.Sprintf("%%%v%%", value),
	}
}

func In[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "IN",
		value:    value,
	}
}

func MultipleIn[T any](field []string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("(`%s`)", strings.Join(field, "`, `")),
		operator: "IN",
		value:    value,
	}
}

func NotIn[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "NOT IN",
		value:    value,
	}
}

func EmptyCriteria() stringCriteria {
	return stringCriteria("")
}

func IsNot[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "!=",
		value:    value,
	}
}

func FindInSet[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "find_in_set",
		value:    value,
	}
}

func FindInSetFormFiled[T any](field string, value T) Criteria {
	return binaryOperatorCriteria[T]{
		field:    fmt.Sprintf("`%s`", field),
		operator: "find_in_set_form_filed",
		value:    value,
	}
}
