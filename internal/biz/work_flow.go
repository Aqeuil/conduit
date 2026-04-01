package biz

type WorkFlow struct {
	FuncKey string                 `json:"func_key"`
	Params  map[string]interface{} `json:"params"`
}
