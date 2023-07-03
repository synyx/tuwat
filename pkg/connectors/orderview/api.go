package orderview

type ticket struct {
	Owner   string `json:"owner"`
	State   string `json:"state"`
	Since   int64  `json:"since"`
	Message string `json:"message"`
}
