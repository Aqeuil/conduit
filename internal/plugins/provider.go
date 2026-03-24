package plugins

import (
	"net/http"
)

type FuncKey string

type PreFunc interface {
	Key() FuncKey
	Execute(req *http.Request, params map[string]any) error
	Help() string
	ParamRules() []ParamRule
}

type PostFunc interface {
	Key() FuncKey
	Execute(req *http.Response, params map[string]any) error
	Help() string
}

type ParamType string

var (
	Number = ParamType("number")
	String = ParamType("string")
	Bool   = ParamType("bool")
	Array  = ParamType("array")
	Object = ParamType("object")
)

type ParamRule struct {
	Type     ParamType
	Name     string
	Children []ParamRule
}

var PrePlugins = make(map[FuncKey]PreFunc)

func RegisterPrePlugin(p PreFunc) {
	PrePlugins[p.Key()] = p
}

var PostPlugins = make(map[FuncKey]PostFunc)

func RegisterPostPlugin(p PostFunc) {
	PostPlugins[p.Key()] = p
}
