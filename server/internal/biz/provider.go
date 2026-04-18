package biz

import (
	"conduit/internal/biz/manager"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	manager.NewManager,
)
