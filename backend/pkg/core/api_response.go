package core

type ApiResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func Success(code int, message string, data interface{}, meta interface{}) *ApiResponse {
	return &ApiResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

func Error(code int, message string, err interface{}, meta interface{}) *ApiResponse {
	return &ApiResponse{
		Success: false,
		Code:    code,
		Message: message,
		Error:   err,
		Meta:    meta,
	}
}
