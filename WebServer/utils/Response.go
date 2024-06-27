package utils

// 不分页响应结构数据
type StructResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func Result(code int, data interface{}, msg string) StructResponse {
	return StructResponse{
		code,
		data,
		msg,
	}
}
