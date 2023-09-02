package orderview

type ticket struct {
	Owner   string `json:"owner"`
	State   int    `json:"state"`
	Since   int64  `json:"since"`
	Message string `json:"message"`
}
