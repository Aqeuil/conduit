package model

import (
	"conduit/internal/data"
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

// Application 应用表
type Application struct {
	tableName   struct{}  `pg:"applications"`
	ID          string    `pg:"id,pk" json:"id"`
	Name        string    `pg:"name,notnull" json:"name"`
	Upstream    string    `pg:"upstream,notnull" json:"upstream"`
	Description string    `pg:"description" json:"description"`
	CreatedAt   time.Time `pg:"created_at,notnull" json:"created_at"`
	UpdatedAt   time.Time `pg:"updated_at,notnull" json:"updated_at"`
}

type ApplicationRepo struct {
	data *data.Data
}

func NewApplicationRepo(data *data.Data) *ApplicationRepo {
	return &ApplicationRepo{data: data}
}

func (r *ApplicationRepo) Create(ctx context.Context, app *Application) error {
	app.ID = uuid.New().String()
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now
	_, err := r.data.Data.ModelContext(ctx, app).Insert()
	return err
}

func (r *ApplicationRepo) Update(ctx context.Context, app *Application) error {
	app.UpdatedAt = time.Now()
	_, err := r.data.Data.ModelContext(ctx, app).
		Column("name", "upstream", "description", "updated_at").
		WherePK().
		Update()
	return err
}

func (r *ApplicationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.data.Data.ModelContext(ctx, &Application{ID: id}).WherePK().Delete()
	return err
}

func (r *ApplicationRepo) Get(ctx context.Context, id string) (*Application, error) {
	app := &Application{ID: id}
	err := r.data.Data.ModelContext(ctx, app).WherePK().Select()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return app, err
}

func (r *ApplicationRepo) List(ctx context.Context, page, pageSize int) ([]*Application, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var apps []*Application
	total, err := r.data.Data.ModelContext(ctx, &apps).
		OrderExpr("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		SelectAndCount()
	return apps, int64(total), err
}
