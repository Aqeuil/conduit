package response

import "time"

type Response[D any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    D      `json:"data"`
	Time    string `json:"time"`
}

func Ok[D any](data D) Response[D] {
	return Response[D]{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

func Fail(message string) Response[any] {
	return Response[any]{
		Code:    500,
		Message: message,
	}
}

func FailWithCode(code int, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Message: message,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}
}
