package types

type Status struct {
	ResultCode    int    `json:"code"`
	ResultMessage string `json:"message"`
}

type ResponseHeader struct {
	Status   Status `json:"status"`
	FromHost string `json:"fromhost"`
}

type RequestHeader struct {
	Status Status `json:"status"`
	ToHost string `json:"tohost"`
}

type Message struct {
	Payload string `json:"payload"`
}

type Request struct {
	Header  RequestHeader `json:"header"`
	Payload string        `json:"payload"`
}

type Response struct {
	Header  ResponseHeader `json:"header"`
	Payload string         `json:"payload"`
}
