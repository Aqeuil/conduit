package criteria

import (
	"fmt"
	"strings"
)

type joinCriteria struct {
	criteriaCollection []Criteria
	separator          string
}

func (s joinCriteria) GetQuery() string {
	queries := make([]string, 0, len(s.criteriaCollection))

	for _, spec := range s.criteriaCollection {
		queries = append(queries, spec.GetQuery())
	}

	return fmt.Sprintf("(%s)", strings.Join(queries, fmt.Sprintf(" %s ", s.separator)))
}

func (s joinCriteria) GetValues() []any {
	values := make([]any, 0)

	for _, spec := range s.criteriaCollection {
		values = append(values, spec.GetValues()...)
	}

	return values
}

//func (s QueryBuilder) Equal(field string, value any) Criteria {
//	return And(s, Binary(field, "=", value))
//}
//
//func (s QueryBuilder) GreaterThan(field string, value any) Criteria {
//	return And(s, Binary(field, ">", value))
//}
//
//func (s QueryBuilder) GreaterOrEqual(field string, value any) Criteria {
//	return And(s, Binary(field, ">=", value))
//}
//
//func (s QueryBuilder) LessThan(field string, value any) Criteria {
//	return And(s, Binary(field, "<", value))
//}
//
//func (s QueryBuilder) LessOrEqual(field string, value any) Criteria {
//	return And(s, Binary(field, "<=", value))
//}
//
//func (s QueryBuilder) LeftLike(field string, value any) Criteria {
//	return And(s, LeftLike(field, value))
//}
//
//func (s QueryBuilder) Like(field string, value any) Criteria {
//	return And(s, Like(field, value))
//}
//
//func (s QueryBuilder) In(field string, value any) Criteria {
//	return And(s, In(field, value))
//}
//
//func (s QueryBuilder) NotIn(field string, value any) Criteria {
//	return And(s, NotIn(field, value))
//}
