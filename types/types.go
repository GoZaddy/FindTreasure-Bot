package types

import (
	"github.com/bot-api/telegram"
)

//TelegramResponse model
type TelegramResponse struct {
	Message  telegram.Message `json:"message"`
	UpdateID int              `json:"update_id"`
}

//EncryptionDetails returns the encryption details for each node
type EncryptionDetails struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
}

//Path defines each hittable path
type Path struct {
	CipherID string `json:"cipherId"`
	Rounds   int    `json:"n"`
}

//Treasures gives treasure details
type Treasures struct {
	Total int `json:"total"`
	Found int `json:"found"`
}

//NodeResponse defines the structure of the response gotten from hitting each node
type NodeResponse struct {
	Encryption EncryptionDetails `json:"encryption"`
	Paths      []Path            `json:"paths"`
	Treasures  Treasures         `json:"treasures"`
}
