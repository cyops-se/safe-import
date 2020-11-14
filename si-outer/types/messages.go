package types

type ByNameRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ByIdRequest struct {
	ID int `json:"id"`
}

type WaitRequest struct {
	URL string `json:"url"`
}

type WaitResponse struct {
	Success  bool   `json:"success"`
	Error    error  `json:"error"`
	Filename string `json:"filename"`
}
