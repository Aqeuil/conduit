package model

import (
	"conduit/internal/data"
	"context"
	"encoding/json"
	"time"

	"github.com/go-pg/pg/v10"
)

const (
	RouterTypeDirectory = "directory"
	RouterTypeRouter    = "router"
)

// Router 统一路由节点，type="directory" 时为目录，type="router" 时为路由
type Router struct {
	tableName      struct{}        `pg:"routers"`
	ID             int             `pg:"id,pk" json:"id"`
	AppID          string          `pg:"app_id" json:"app_id"`
	ParentID       int             `pg:"parent_id" json:"parent_id"`
	Type           string          `pg:"type" json:"type"`
	Name           string          `pg:"name" json:"name"`
	SortOrder      int             `pg:"sort_order" json:"sort_order"`
	Path           string          `pg:"path" json:"path"`
	Method         string          `pg:"method" json:"method"`
	Headers        json.RawMessage `pg:"headers,type:jsonb" json:"headers"`
	RequestSchema  json.RawMessage `pg:"request_schema,type:jsonb" json:"request_schema"`
	ResponseSchema json.RawMessage `pg:"response_schema,type:jsonb" json:"response_schema"`
	Description    string          `pg:"description" json:"description"`
	CreatedAt      time.Time       `pg:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `pg:"updated_at" json:"updated_at"`
}

// RouterRepo CRUD for unified routers table
type RouterRepo struct {
	data *data.Data
}

func NewRouterRepo(data *data.Data) *RouterRepo {
	return &RouterRepo{data: data}
}

func (r *RouterRepo) Create(ctx context.Context, rt *Router) error {
	now := time.Now()
	rt.CreatedAt = now
	rt.UpdatedAt = now
	_, err := r.data.Data.ModelContext(ctx, rt).Insert()
	return err
}

func (r *RouterRepo) Update(ctx context.Context, rt *Router) error {
	rt.UpdatedAt = time.Now()
	_, err := r.data.Data.ModelContext(ctx, rt).
		Column("parent_id", "name", "sort_order", "path", "method",
			"headers", "request_schema", "response_schema", "description", "updated_at").
		WherePK().
		Update()
	return err
}

func (r *RouterRepo) Delete(ctx context.Context, id int) error {
	_, err := r.data.Data.ModelContext(ctx, &Router{ID: id}).WherePK().Delete()
	return err
}

func (r *RouterRepo) Get(ctx context.Context, id int) (*Router, error) {
	rt := &Router{ID: id}
	err := r.data.Data.ModelContext(ctx, rt).WherePK().Select()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return rt, err
}

func (r *RouterRepo) ListAll(ctx context.Context, appID string) ([]*Router, error) {
	var list []*Router
	err := r.data.Data.ModelContext(ctx, &list).
		Where("app_id = ?", appID).
		OrderExpr("type DESC, sort_order, id").
		Select()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return list, err
}
