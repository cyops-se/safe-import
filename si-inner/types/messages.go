package types

type ByNameRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ByIdRequest struct {
	ID uint `json:"id"`
}

type ApproveRequest struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type HttpClassificationRequest struct {
	ID    uint   `json:"id"`
	Class string `json:"string"`
	Allow bool   `json:"allow"`
}
