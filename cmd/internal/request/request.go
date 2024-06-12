package request

type Request struct {
	Function string                  `json:"function"`
	Args     *map[string]interface{} `json:"args"`
}
