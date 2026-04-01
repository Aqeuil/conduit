package service

import (
	"conduit/internal/service/plugin"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewConduitServer,
	plugin.NewPluginServer,
	plugin.NewPluginAdminServer,
)
