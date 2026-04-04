package model

import (
	"conduit/internal/data"
	"context"
	"encoding/json"
	"time"

	"github.com/go-pg/pg/v10"
)

// Workflow Application的工作流表
type Workflow struct {
	tableName    struct{}        `pg:"workflows"`
	ID           int             `pg:"id,pk" json:"id"`
	AppID        string          `pg:"app_id,notnull" json:"app_id"`
	WorkflowType string          `pg:"workflow_type,notnull" json:"workflow_type"`
	FuncKey      string          `pg:"func_key,notnull" json:"func_key"`
	Params       json.RawMessage `pg:"params,type:jsonb" json:"params"`
	SortOrder    int             `pg:"sort_order,notnull" json:"sort_order"`
	CreatedAt    time.Time       `pg:"created_at,notnull" json:"created_at"`
	UpdatedAt    time.Time       `pg:"updated_at,notnull" json:"updated_at"`
}

// ---- WorkflowRepo ----

type WorkflowRepo struct {
	data *data.Data
}

func NewWorkflowRepo(data *data.Data) *WorkflowRepo {
	return &WorkflowRepo{data: data}
}

func (r *WorkflowRepo) Create(ctx context.Context, w *Workflow) error {
	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now
	_, err := r.data.Data.ModelContext(ctx, w).Insert()
	return err
}

func (r *WorkflowRepo) Update(ctx context.Context, w *Workflow) error {
	w.UpdatedAt = time.Now()
	_, err := r.data.Data.ModelContext(ctx, w).
		Column("workflow_type", "func_key", "params", "sort_order", "updated_at").
		WherePK().
		Update()
	return err
}

func (r *WorkflowRepo) Delete(ctx context.Context, id int) error {
	_, err := r.data.Data.ModelContext(ctx, &Workflow{ID: id}).WherePK().Delete()
	return err
}

func (r *WorkflowRepo) List(ctx context.Context, appID, workflowType string) ([]*Workflow, error) {
	var workflows []*Workflow
	q := r.data.Data.ModelContext(ctx, &workflows).
		Where("app_id = ?", appID).
		OrderExpr("sort_order, id")

	if workflowType != "" {
		q = q.Where("workflow_type = ?", workflowType)
	}

	err := q.Select()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return workflows, err
}
