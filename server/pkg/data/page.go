package data

import (
	"fmt"
	"strings"
)

// Page 分页信息
type Page[E any] struct {
	FirstPage  int //指定第一页的页码id，默认是从1开始
	TotalPage  int
	TotalItems int64
	PageSize   int
	Page       int
	Pageable   Pageable
	Items      []E
}

func NewPage[E any]() *Page[E] {
	return &Page[E]{
		FirstPage:  1,
		TotalPage:  0,
		TotalItems: 0,
		PageSize:   0,
		Items:      []E{},
		Pageable:   nil,
	}
}

func NewPageWith[E any](totalItems int64, pageSize int, page int, items []E, pageable Pageable) *Page[E] {
	totalPage := 0
	if pageSize > 0 {
		totalPage := int(totalItems / int64(pageSize))
		if totalItems%int64(pageSize) != 0 {
			totalPage += 1
		}
	}

	return &Page[E]{
		FirstPage:  1,
		TotalPage:  totalPage,
		TotalItems: totalItems,
		PageSize:   pageSize,
		Page:       page,
		Items:      items,
		Pageable:   pageable,
	}
}

func (p *Page[E]) IsEmpty() bool {
	return len(p.Items) == 0
}

func (p *Page[E]) HasNext() bool {
	return p.Page < p.TotalPage+p.FirstPage-1
}

func (p *Page[E]) HasPrevious() bool {
	return p.Page > p.FirstPage
}

func (p *Page[E]) NextPageable() Pageable {
	if p.Pageable == nil {
		return nil
	}
	return p.Pageable.Next()
}

func (p *Page[E]) PreviousPageable() Pageable {
	if p.Pageable == nil {
		return nil
	}
	return p.Pageable.PreviousOrFirst()
}

// Pageable 分页请求接口
type Pageable interface {
	GetPageNumber() int
	GetPageSize() int
	GetOffset() int
	GetSort() *Sort
	GetFirstPage() int
	HasPrevious() bool
	First() Pageable
	Next() Pageable
	PreviousOrFirst() Pageable
	WithPage(pageNumber int) Pageable
	WithPageSize(pageSize int) Pageable
	WithSort(sort *Sort)
}

// Pagination 实现Pageable接口，代表分页请求信息
type Pagination struct {
	Page      int
	PageSize  int
	FirstPage int
	Sort      *Sort
}

func NewPagination(page int, pageSize int, sort *Sort) *Pagination {
	if sort == nil {
		sort = NewSort()
	}
	return &Pagination{
		Page:      page,
		PageSize:  pageSize,
		FirstPage: 1,
		Sort:      sort,
	}
}

func (p *Pagination) GetPageNumber() int {
	return p.Page
}

func (p *Pagination) GetPageSize() int {
	return p.PageSize
}

func (p *Pagination) GetSort() *Sort {
	if p.Sort == nil {
		p.Sort = NewSort()
	}
	return p.Sort
}

func (p *Pagination) GetOffset() int {
	return (p.Page - p.FirstPage) * p.PageSize
}

func (p *Pagination) GetFirstPage() int {
	return p.FirstPage
}

func (p *Pagination) HasPrevious() bool {
	return p.Page > p.FirstPage
}

func (p *Pagination) First() Pageable {
	return NewPagination(p.FirstPage, p.PageSize, p.Sort)
}

func (p *Pagination) Next() Pageable {
	return NewPagination(p.Page+1, p.PageSize, p.Sort)
}

func (p *Pagination) PreviousOrFirst() Pageable {
	if p.HasPrevious() {
		return NewPagination(p.Page, p.PageSize, p.Sort)
	} else {
		return p.First()
	}
}

func (p *Pagination) WithPage(pageNumber int) Pageable {
	p.Page = pageNumber
	return p
}

func (p *Pagination) WithPageSize(pageSize int) Pageable {
	p.PageSize = pageSize
	return p
}

func (p *Pagination) WithSort(sort *Sort) {
	p.GetSort().And(sort)
}

func (p *Pagination) WithSortDesc(property string) {
	p.GetSort().OrderDesc(property)
}

func (p *Pagination) WithSortAsc(property string) {
	p.GetSort().OrderAsc(property)
}

type OrderDirection int

const (
	AscOrder  OrderDirection = 1
	DescOrder OrderDirection = 2
)

type StringSort string

func (s StringSort) GetValue() string {
	return string(s)
}

// Sort 查询数据的排序信息
type Sort struct {
	Orders map[string]OrderDirection
}

func NewSort() *Sort {
	sort := &Sort{
		Orders: map[string]OrderDirection{},
	}
	return sort
}

func NewSortOrder(property string, direction OrderDirection) *Sort {
	sort := &Sort{
		Orders: map[string]OrderDirection{},
	}
	sort.Orders[property] = direction
	return sort
}

func (s *Sort) HasSort() bool {
	return s.Orders != nil && len(s.Orders) > 0
}

func (s *Sort) OrderDesc(property string) *Sort {
	return s.OrderBy(property, DescOrder)
}

func (s *Sort) OrderAsc(property string) *Sort {
	return s.OrderBy(property, AscOrder)
}

func (s *Sort) OrderBy(property string, direction OrderDirection) *Sort {
	if s.Orders == nil {
		s.Orders = map[string]OrderDirection{}
	}
	s.Orders[property] = direction
	return s
}

func (s *Sort) And(sort *Sort) *Sort {
	if s == nil {
		return sort
	}

	if s.Orders == nil {
		s.Orders = map[string]OrderDirection{}
	}
	if sort != nil {
		for property, direction := range sort.Orders {
			s.Orders[property] = direction
		}
	}
	return s
}

func (s *Sort) GetValue() string {
	if s == nil || len(s.Orders) == 0 {
		return ""
	}

	var orders []string
	for property, direction := range s.Orders {
		var order string
		if direction == AscOrder {
			order = fmt.Sprintf("%s", property)
		} else {
			order = fmt.Sprintf("%s DESC", property)
		}
		orders = append(orders, order)
	}

	return strings.Join(orders, ",")
}
