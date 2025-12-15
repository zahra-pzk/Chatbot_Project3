package dto

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
