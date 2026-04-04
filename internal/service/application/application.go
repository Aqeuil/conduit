package application

import (
	admin "conduit/api/v1/conduit-admin"
	"conduit/internal/model"
	"context"
	"time"
)

type ApplicationAdminServer struct {
	admin.UnimplementedApplicationAdminServer
	repo *model.ApplicationRepo
}

func NewApplicationAdminServer(repo *model.ApplicationRepo) *ApplicationAdminServer {
	return &ApplicationAdminServer{repo: repo}
}

func (s *ApplicationAdminServer) CreateApplication(ctx context.Context, req *admin.CreateApplicationReq) (*admin.ApplicationInfo, error) {
	app := &model.Application{
		Name:        req.Name,
		Upstream:    req.Upstream,
		Description: req.Description,
	}
	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}
	return toInfo(app), nil
}

func (s *ApplicationAdminServer) UpdateApplication(ctx context.Context, req *admin.UpdateApplicationReq) (*admin.ApplicationInfo, error) {
	app := &model.Application{
		ID:          req.Id,
		Name:        req.Name,
		Upstream:    req.Upstream,
		Description: req.Description,
	}
	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}
	return toInfo(app), nil
}

func (s *ApplicationAdminServer) DeleteApplication(ctx context.Context, req *admin.DeleteApplicationReq) (*admin.DeleteApplicationResp, error) {
	if err := s.repo.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &admin.DeleteApplicationResp{}, nil
}

func (s *ApplicationAdminServer) GetApplication(ctx context.Context, req *admin.GetApplicationReq) (*admin.ApplicationInfo, error) {
	app, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toInfo(app), nil
}

func (s *ApplicationAdminServer) ListApplications(ctx context.Context, req *admin.ListApplicationsReq) (*admin.ListApplicationsResp, error) {
	apps, total, err := s.repo.List(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	resp := &admin.ListApplicationsResp{Total: total}
	for _, app := range apps {
		resp.Items = append(resp.Items, toInfo(app))
	}
	return resp, nil
}

func toInfo(app *model.Application) *admin.ApplicationInfo {
	return &admin.ApplicationInfo{
		Id:          app.ID,
		Name:        app.Name,
		Upstream:    app.Upstream,
		Description: app.Description,
		CreatedAt:   app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   app.UpdatedAt.Format(time.RFC3339),
	}
}
