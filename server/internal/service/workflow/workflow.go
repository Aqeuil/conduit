package workflow

import (
	admin "conduit/api/v1/conduit-admin"
	"conduit/internal/model"
	"context"
	"time"
)

type WorkflowAdminServer struct {
	admin.UnimplementedWorkflowAdminServer
	repo *model.WorkflowRepo
}

func NewWorkflowAdminServer(repo *model.WorkflowRepo) *WorkflowAdminServer {
	return &WorkflowAdminServer{repo: repo}
}

func (s *WorkflowAdminServer) CreateWorkflow(ctx context.Context, req *admin.CreateWorkflowReq) (*admin.WorkflowInfo, error) {
	w := &model.Workflow{
		AppID:        req.AppId,
		WorkflowType: req.WorkflowType,
		FuncKey:      req.FuncKey,
		Params:       []byte(req.Params),
		SortOrder:    int(req.SortOrder),
	}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return toInfo(w), nil
}

func (s *WorkflowAdminServer) UpdateWorkflow(ctx context.Context, req *admin.UpdateWorkflowReq) (*admin.WorkflowInfo, error) {
	w := &model.Workflow{
		ID:           int(req.Id),
		WorkflowType: req.WorkflowType,
		FuncKey:      req.FuncKey,
		Params:       []byte(req.Params),
		SortOrder:    int(req.SortOrder),
	}
	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}
	return toInfo(w), nil
}

func (s *WorkflowAdminServer) DeleteWorkflow(ctx context.Context, req *admin.DeleteWorkflowReq) (*admin.DeleteWorkflowResp, error) {
	if err := s.repo.Delete(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &admin.DeleteWorkflowResp{}, nil
}

func (s *WorkflowAdminServer) ListWorkflows(ctx context.Context, req *admin.ListWorkflowsReq) (*admin.ListWorkflowsResp, error) {
	workflows, err := s.repo.List(ctx, req.AppId, req.WorkflowType)
	if err != nil {
		return nil, err
	}
	resp := &admin.ListWorkflowsResp{}
	for _, w := range workflows {
		resp.Items = append(resp.Items, toInfo(w))
	}
	return resp, nil
}

func toInfo(w *model.Workflow) *admin.WorkflowInfo {
	return &admin.WorkflowInfo{
		Id:           int32(w.ID),
		AppId:        w.AppID,
		WorkflowType: w.WorkflowType,
		FuncKey:      w.FuncKey,
		Params:       string(w.Params),
		SortOrder:    int32(w.SortOrder),
		CreatedAt:    w.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    w.UpdatedAt.Format(time.RFC3339),
	}
}
