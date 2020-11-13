package types

type ByNameRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ByIdRequest struct {
	ID int `json:"id"`
}

type WaitRequest struct {
	RepositoryID int    `json:"repoid"`
	URL          string `json:"url"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type WaitResponse struct {
	Success  bool   `json:"success"`
	Error    error  `json:"error"`
	Filename string `json:"filename"`
}
