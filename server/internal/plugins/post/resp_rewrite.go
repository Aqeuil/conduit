package post

import (
	"bytes"
	"conduit/internal/plugins"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type RespRewrite struct {
}

func (r RespRewrite) Key() plugins.FuncKey {
	return "resp_rewrite"
}

func (r RespRewrite) Execute(req *http.Request, resp *http.Response, params map[string]any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		// If not JSON, treat as string
		data = string(body)
	}

	code := 0
	msg := "success"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code = resp.StatusCode
		msg = http.StatusText(resp.StatusCode)
	}

	result := map[string]any{
		"code": code,
		"msg":  msg,
		"data": data,
	}

	newBody, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(newBody))
	resp.ContentLength = int64(len(newBody))
	resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
	resp.Header.Set("Content-Type", "application/json")

	return nil
}

func (r RespRewrite) Help() string {
	return "重写响应体, 使用 code, msg, data 进行包裹"
}

func (r RespRewrite) ParamRules() []plugins.ParamRule {
	return nil
}

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

type ListNode struct {
	Val  int
	Next *ListNode
}

func rotateRight(head *ListNode, k int) *ListNode {
	length := 0
	p := head
	for true {
		length++
		if p.Next == nil {
			p.Next = head
			break
		}

		p = p.Next
	}

	k = k % length
	target := length - k
	res := head
	for idx := 0; idx < target-1; idx++ {
		res = res.Next
	}
	fmt.Println(res.Val)

	next := res.Next
	res.Next = nil
	return next
}
