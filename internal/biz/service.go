package biz

import "net/http"

// ServiceUnit 服务单元, 最小的Deployment单位
type ServiceUnit struct {
	// Id DeploymentId
	Id string

	// Host 请求地址
	Host string
}

func (s ServiceUnit) PreProgress(req *http.Request) error {
	return nil
}

func (s ServiceUnit) PostProgress(req *http.Response) error {
	return nil
}
