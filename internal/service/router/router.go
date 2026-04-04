package router

import (
	admin "conduit/api/v1/conduit-admin"
	"conduit/internal/model"
	"context"
	"time"
)

type RouterAdminServer struct {
	admin.UnimplementedRouterAdminServer
	repo *model.RouterRepo
}

func NewRouterAdminServer(repo *model.RouterRepo) *RouterAdminServer {
	return &RouterAdminServer{repo: repo}
}

func (s *RouterAdminServer) CreateRouter(ctx context.Context, req *admin.CreateRouterReq) (*admin.RouterNode, error) {
	rt := &model.Router{
		AppID:          req.AppId,
		ParentID:       int(req.ParentId),
		Type:           req.Type,
		Name:           req.Name,
		SortOrder:      int(req.SortOrder),
		Path:           req.Path,
		Method:         req.Method,
		Headers:        jsonOrDefault(req.Headers, "[]"),
		RequestSchema:  jsonOrDefault(req.RequestSchema, "{}"),
		ResponseSchema: jsonOrDefault(req.ResponseSchema, "{}"),
		Description:    req.Description,
	}
	if err := s.repo.Create(ctx, rt); err != nil {
		return nil, err
	}
	return nodeToProto(rt), nil
}

func (s *RouterAdminServer) UpdateRouter(ctx context.Context, req *admin.UpdateRouterReq) (*admin.RouterNode, error) {
	rt := &model.Router{
		ID:             int(req.Id),
		ParentID:       int(req.ParentId),
		Name:           req.Name,
		SortOrder:      int(req.SortOrder),
		Path:           req.Path,
		Method:         req.Method,
		Headers:        jsonOrDefault(req.Headers, "[]"),
		RequestSchema:  jsonOrDefault(req.RequestSchema, "{}"),
		ResponseSchema: jsonOrDefault(req.ResponseSchema, "{}"),
		Description:    req.Description,
	}
	if err := s.repo.Update(ctx, rt); err != nil {
		return nil, err
	}
	return nodeToProto(rt), nil
}

func (s *RouterAdminServer) DeleteRouter(ctx context.Context, req *admin.DeleteRouterReq) (*admin.DeleteRouterResp, error) {
	if err := s.repo.Delete(ctx, int(req.Id)); err != nil {
		return nil, err
	}
	return &admin.DeleteRouterResp{}, nil
}

func (s *RouterAdminServer) GetRouter(ctx context.Context, req *admin.GetRouterReq) (*admin.RouterNode, error) {
	rt, err := s.repo.Get(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return nodeToProto(rt), nil
}

func (s *RouterAdminServer) ListRouters(ctx context.Context, req *admin.ListRoutersReq) (*admin.ListRoutersResp, error) {
	all, err := s.repo.ListAll(ctx, req.AppId)
	if err != nil {
		return nil, err
	}
	return &admin.ListRoutersResp{Routers: buildTree(all)}, nil
}

// buildTree builds a recursive tree from a flat list ordered by type DESC (directories first).
func buildTree(nodes []*model.Router) []*admin.RouterNode {
	protoMap := make(map[int]*admin.RouterNode, len(nodes))
	for _, n := range nodes {
		protoMap[n.ID] = nodeToProto(n)
	}

	var roots []*admin.RouterNode
	for _, n := range nodes {
		p := protoMap[n.ID]
		if n.ParentID == 0 {
			roots = append(roots, p)
		} else if parent, ok := protoMap[n.ParentID]; ok {
			parent.Children = append(parent.Children, p)
		} else {
			// orphan: parent deleted; treat as root
			roots = append(roots, p)
		}
	}
	return roots
}

func nodeToProto(rt *model.Router) *admin.RouterNode {
	return &admin.RouterNode{
		Id:             int32(rt.ID),
		AppId:          rt.AppID,
		ParentId:       int32(rt.ParentID),
		Type:           rt.Type,
		Name:           rt.Name,
		SortOrder:      int32(rt.SortOrder),
		Path:           rt.Path,
		Method:         rt.Method,
		Headers:        string(rt.Headers),
		RequestSchema:  string(rt.RequestSchema),
		ResponseSchema: string(rt.ResponseSchema),
		Description:    rt.Description,
		CreatedAt:      rt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      rt.UpdatedAt.Format(time.RFC3339),
	}
}

func jsonOrDefault(s, def string) []byte {
	if s == "" {
		return []byte(def)
	}
	return []byte(s)
}
