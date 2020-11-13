package types

type ChatMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}
