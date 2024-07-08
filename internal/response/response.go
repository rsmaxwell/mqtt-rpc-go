package response

import (
	"fmt"
	"net/http"
)

type Response struct {
	Data map[string]interface{} `json:"data"`
}

func New() *Response {
	var r Response
	r.Data = make(map[string]interface{})
	return &r
}

func Success() *Response {
	r := New()
	r.Data["code"] = http.StatusOK
	return r
}

func BadRequest() *Response {
	r := New()
	r.Data["code"] = http.StatusBadRequest
	return r
}

func BadInternalServerError() *Response {
	r := New()
	r.Data["code"] = http.StatusInternalServerError
	return r
}

func (r Response) Ok() bool {
	code, err := r.GetCode()
	if err != nil {
		return false
	}
	return code == http.StatusOK
}

func (r Response) PutCode(code int) {
	r.Data["code"] = code
}

func (r Response) GetCode() (int, error) {
	value := r.Data["code"]
	v, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("unexpected type for '%s': %+v", "code", value)
	}
	return v, nil
}

func (r Response) PutMessage(message string) {
	r.Data["message"] = message
}

func (r Response) GetMessage() (string, error) {
	return r.GetString("message")
}

func (r Response) PutString(key string, value string) {
	r.Data[key] = value
}

func (r Response) GetString(key string) (string, error) {
	value := r.Data[key]
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Response) PutInteger(key string, value int64) {
	r.Data[key] = value
}

func (r Response) GetInteger(key string) (int64, error) {
	v, err := r.GetNumber(key)
	return int64(v), err
}

func (r Response) PutNumber(key string, value float64) {
	r.Data[key] = value
}

func (r Response) GetNumber(key string) (float64, error) {
	value := r.Data[key]
	v, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Response) PutBoolean(key string, value bool) {
	r.Data[key] = value
}

func (r Response) GetBoolean(key string) (bool, error) {
	value := r.Data[key]
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}
