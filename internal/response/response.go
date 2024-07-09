package response

import (
	"fmt"
	"math"
	"net/http"
)

type Response map[string]interface{}

func New() *Response {
	r := make(Response)
	return &r
}

func Success() *Response {
	r := New()
	(*r)["code"] = http.StatusOK
	return r
}

func BadRequest() *Response {
	r := New()
	(*r)["code"] = http.StatusBadRequest
	return r
}

func BadInternalServerError() *Response {
	r := New()
	(*r)["code"] = http.StatusInternalServerError
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
	r["code"] = code
}

func (r Response) GetCode() (int, error) {
	value, err := r.GetInteger("code")
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func (r Response) PutMessage(message string) {
	r["message"] = message
}

func (r Response) GetMessage() (string, error) {
	return r.GetString("message")
}

func (r Response) PutString(key string, value string) {
	r[key] = value
}

func (r Response) GetString(key string) (string, error) {
	value := r[key]
	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Response) PutInteger(key string, value int64) {
	r[key] = value
}

func (r Response) GetInteger(key string) (int64, error) {
	v, err := r.GetNumber(key)
	return int64(math.Round(v)), err
}

func (r Response) PutNumber(key string, value float64) {
	r[key] = value
}

func (r Response) GetNumber(key string) (float64, error) {
	value := r[key]
	v, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}

func (r Response) PutBoolean(key string, value bool) {
	r[key] = value
}

func (r Response) GetBoolean(key string) (bool, error) {
	value := r[key]
	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for '%s': %+v", key, value)
	}
	return v, nil
}
