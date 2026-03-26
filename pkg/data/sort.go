package data

import (
	"fmt"
	"strings"
)

// 查询数据的排序信息
type OrderByColumn struct {
	Column    string
	Direction OrderDirection
}

type OrderBy struct {
	Columns []OrderByColumn
}

func NewOrderBy(column string, direction OrderDirection) *OrderBy {
	orderBy := &OrderBy{
		Columns: []OrderByColumn{{Column: column, Direction: direction}},
	}
	return orderBy
}

func (s *OrderBy) HasSort() bool {
	return s.Columns != nil && len(s.Columns) > 0
}

func (s *OrderBy) OrderDesc(column string) *OrderBy {
	return s.OrderBy(column, DescOrder)
}

func (s *OrderBy) OrderAsc(column string) *OrderBy {
	return s.OrderBy(column, AscOrder)
}

func (s *OrderBy) OrderBy(column string, direction OrderDirection) *OrderBy {
	if s.Columns == nil {
		s.Columns = []OrderByColumn{}
	}
	s.Columns = append(s.Columns, OrderByColumn{Column: column, Direction: direction})
	return s
}

func (s *OrderBy) GetValue() string {
	if s == nil || len(s.Columns) == 0 {
		return ""
	}

	var orders []string
	for _, order := range s.Columns {
		var direction string
		if order.Direction == AscOrder {
			direction = ""
		} else {
			direction = " DESC"
		}
		orders = append(orders, fmt.Sprintf("%s%s", order.Column, direction))
	}

	return strings.Join(orders, ",")
}
