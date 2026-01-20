package conduit

import (
	"go.conduit.cn/conduit/protocols/v2"
	"go.conduit.cn/conduit/v2/server/config"
)

func Main() {
	// -1. 解析配置
	err := config.Parse()
	if err != nil {
		panic(err)
	}

	protocols.StartListener()
}
