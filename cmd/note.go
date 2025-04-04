package main

import (
	"encoding/hex"
	"log"
)

// Note - структура для хранения заметок.
type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Login string `json:"login"`
	Pass  string `json:"pass"` // хранится в hex-кодировке
}

// EncryptPassword шифрует пароль и сохраняет его в hex-формате.
func (n *Note) EncryptPassword(plainPass string) {
	if plainPass == "" {
		n.Pass = ""
		return
	}
	enc, err := encrypt([]byte(plainPass), encryptionKey)
	if err != nil {
		log.Println("Ошибка шифрования пароля:", err)
		n.Pass = ""
		return
	}
	n.Pass = hex.EncodeToString(enc)
}

// DecryptPassword расшифровывает пароль из hex-формата.
func (n *Note) DecryptPassword() string {
	if n.Pass == "" {
		return ""
	}
	encBytes, err := hex.DecodeString(n.Pass)
	if err != nil {
		log.Println("Ошибка hex-декодирования:", err)
		return ""
	}
	dec, err := decrypt(encBytes, encryptionKey)
	if err != nil {
		log.Println("Ошибка расшифрования:", err)
		return ""
	}
	return string(dec)
}
