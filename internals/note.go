package main

import (
	"encoding/hex"
	"log"
)

// Note описывает одну «заметку» с логином и паролем
type Note struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Login    string `json:"login"`
	Pass     string `json:"pass"`
	Favorite bool   `json:"favorite"`
}

// EncryptPassword шифрует поле пароля
func (n *Note) EncryptPassword(plainPass string) {
	if plainPass == "" {
		n.Pass = ""
		return
	}
	enc, err := encrypt([]byte(plainPass), encryptionKey)
	if err != nil {
		log.Println(redText("Ошибка шифрования пароля:"), err)
		n.Pass = ""
		return
	}
	n.Pass = hex.EncodeToString(enc)
}

// DecryptPassword расшифровывает поле пароля
func (n *Note) DecryptPassword() string {
	if n.Pass == "" {
		return ""
	}
	encBytes, err := hex.DecodeString(n.Pass)
	if err != nil {
		log.Println(redText("Ошибка hex-декодирования:"), err)
		return ""
	}
	dec, err := decrypt(encBytes, encryptionKey)
	if err != nil {
		log.Println(redText("Ошибка расшифрования:"), err)
		return ""
	}
	return string(dec)
}

// EncryptTitle шифрует заголовок (Title)
func (n *Note) EncryptTitle() {
	if n.Title == "" {
		return
	}
	enc, err := encrypt([]byte(n.Title), encryptionKey)
	if err != nil {
		log.Println(redText("Ошибка шифрования заголовка:"), err)
		return
	}
	n.Title = hex.EncodeToString(enc)
}

// DecryptTitle расшифровывает поле заголовка (Title)
func (n *Note) DecryptTitle() string {
	if !isHex(n.Title) {
		return n.Title
	}
	encBytes, err := hex.DecodeString(n.Title)
	if err != nil {
		log.Println(redText("Ошибка hex-декодирования заголовка:"), err)
		return n.Title
	}
	dec, err := decrypt(encBytes, encryptionKey)
	if err != nil {
		log.Println(redText("Ошибка расшифрования заголовка:"), err)
		return n.Title
	}
	return string(dec)
}

// isHex проверяет, является ли строка корректным hex
func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
