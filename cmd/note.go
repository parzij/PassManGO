package main

import (
	"encoding/hex"
	"log"
)

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Login string `json:"login"`
	Pass  string `json:"pass"`
}

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
