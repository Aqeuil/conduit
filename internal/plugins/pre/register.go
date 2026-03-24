package pre

import "conduit/internal/plugins"

func init() {
	plugins.RegisterPrePlugin(&Redirect{})
}
