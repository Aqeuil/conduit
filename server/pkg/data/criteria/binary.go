package criteria

import (
	"fmt"
)

type binaryOperatorCriteria[T any] struct {
	field    string
	operator string
	value    T
}

func (s binaryOperatorCriteria[T]) GetQuery() string {
	if s.operator == "IN" {
		return fmt.Sprintf("%s %s (?)", s.field, s.operator)
	}
	if s.operator == "find_in_set" {
		return fmt.Sprintf("find_in_set(%s, ?)", s.field)
	}
	if s.operator == "find_in_set_form_filed" {
		return fmt.Sprintf("find_in_set(?, %s)", s.field)
	}
	return fmt.Sprintf("%s %s ?", s.field, s.operator)
}

func (s binaryOperatorCriteria[T]) GetValues() []any {
	return []any{s.value}
}
