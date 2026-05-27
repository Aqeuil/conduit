package post

import (
	"bytes"
	"conduit/internal/plugins"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type RespChecker struct {
}

func (r RespChecker) Key() plugins.FuncKey {
	return "resp_checker"
}

func (r RespChecker) Execute(req *http.Request, resp *http.Response, params map[string]any) error {
	codeField, _ := params["code_field"].(string)
	successCode := params["success_code"]
	msgField, _ := params["msg_field"].(string)

	if codeField == "" {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Restore body for other plugins or the client
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil // Not JSON, skip checking
	}

	val, ok := data[codeField]
	if !ok {
		return nil
	}

	// Compare success code. JSON unmarshals numbers to float64.
	isSuccess := false
	switch s := successCode.(type) {
	case string:
		isSuccess = fmt.Sprintf("%v", val) == s
	case float64:
		if v, ok := val.(float64); ok {
			isSuccess = v == s
		} else {
			isSuccess = fmt.Sprintf("%v", val) == fmt.Sprintf("%v", s)
		}
	default:
		isSuccess = fmt.Sprintf("%v", val) == fmt.Sprintf("%v", successCode)
	}

	if !isSuccess {
		errMsg := fmt.Sprintf("Third party error: %v", val)
		if msgField != "" {
			if m, ok := data[msgField].(string); ok && m != "" {
				errMsg = m
			}
		}
		return errors.New(errMsg)
	}

	return nil
}

func (r RespChecker) Help() string {
	return "校验三方响应体中的业务状态码"
}

func (r RespChecker) ParamRules() []plugins.ParamRule {
	return []plugins.ParamRule{
		{Type: plugins.String, Name: "code_field"},
		{Type: plugins.String, Name: "success_code"},
		{Type: plugins.String, Name: "msg_field"},
	}
}
