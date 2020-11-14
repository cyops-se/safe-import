package types

type ByNameRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ByIdRequest struct {
	ID uint `json:"id"`
}

type HttpDownloadRequest struct {
	URL string `json:"url"`
}

type HttpDownloadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}
