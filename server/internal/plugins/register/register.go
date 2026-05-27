package register

import (
	"conduit/internal/plugins"
	"conduit/internal/plugins/post"
	"conduit/internal/plugins/pre"
)

func init() {
	// Pre Plugins
	plugins.RegisterPrePlugin(&pre.Redirect{})
	plugins.RegisterPrePlugin(&pre.IpLocation{})

	// Post Plugins
	plugins.RegisterPostPlugin(&post.RespHeader{})
	plugins.RegisterPostPlugin(&post.RespRewrite{})
	plugins.RegisterPostPlugin(&post.RespChecker{})
	plugins.RegisterPostPlugin(post.NewAccessLog())
}
