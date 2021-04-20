package types

type ByNameRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ByIdRequest struct {
	ID uint `json:"id"`
}

type NameValue struct {
	Name  string `json:"Name"`
	Value string `json:"value"`
}

type HttpDownloadRequest struct {
	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Headers []NameValue `json:"headers"`
	Body    string      `json:"body"`
}

type HttpDownloadResponse struct {
	URL      string      `json:"url"`
	Filename string      `json:"filename"`
	Headers  []NameValue `json:"headers"`
}
