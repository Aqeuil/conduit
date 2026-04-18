package register

import (
	"conduit/internal/plugins"
	"conduit/internal/plugins/pre"
)

func init() {
	plugins.RegisterPrePlugin(&pre.Redirect{})
}
