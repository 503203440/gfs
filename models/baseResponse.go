package models

type BaseResponse struct {
	Code string `json:code`
	Msg  string `json:msg`
	Data any    `json:data`
}

func ApiSuccess(data any) BaseResponse {
	return BaseResponse{
		Code: "0",
		Data: data,
	}
}

func ApiError(msg string) BaseResponse {
	return BaseResponse{
		Code: "999",
		Msg:  msg,
	}
}

func ApiErrorCode(msg, code string) BaseResponse {
	return BaseResponse{
		Msg:  msg,
		Code: code,
	}
}

func ApiErrorDetail(msg, code string, data any) BaseResponse {
	return BaseResponse{
		Msg:  msg,
		Code: code,
		Data: data,
	}
}
