package service

import (
	"conduit/internal/service/application"
	"conduit/internal/service/plugin"
	"conduit/internal/service/router"
	"conduit/internal/service/workflow"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewConduitServer,
	plugin.NewPluginServer,
	plugin.NewPluginAdminServer,
	application.NewApplicationAdminServer,
	router.NewRouterAdminServer,
	workflow.NewWorkflowAdminServer,
)
