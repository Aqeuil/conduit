package biz

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
)

type ConduitServer struct {
	log *log.Helper
}

func NewConduitServer(logger log.Logger) *ConduitServer {
	return &ConduitServer{
		log: log.NewHelper(logger),
	}
}

func (c ConduitServer) ServeHTTP(http.ResponseWriter, *http.Request) {

}
