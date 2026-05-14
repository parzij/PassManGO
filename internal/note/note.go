package note

import (
	"encoding/hex"
	"log"

	"parzij/PassManGO/internal/colors"
	"parzij/PassManGO/internal/crypto"
)

type Note struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Login    string `json:"login"`
	Pass     string `json:"pass"`
	Favorite bool   `json:"favorite"`
}

var encryptionKey []byte

// SetEncryptionKey устанавливает ключ шифрования (вызывается из main)
func SetEncryptionKey(key []byte) {
	encryptionKey = key
}

func (n *Note) EncryptPassword(plainPass string) {
	if plainPass == "" {
		n.Pass = ""
		return
	}
	enc, err := crypto.Encrypt([]byte(plainPass), encryptionKey)
	if err != nil {
		log.Println(colors.RedText("Ошибка шифрования пароля:"), err)
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
		log.Println(colors.RedText("Ошибка hex-декодирования пароля:"), err)
		return ""
	}
	dec, err := crypto.Decrypt(encBytes, encryptionKey)
	if err != nil {
		log.Println(colors.RedText("Ошибка расшифровки пароля:"), err)
		return ""
	}
	return string(dec)
}

func (n *Note) EncryptTitle() {
	if n.Title == "" {
		return
	}
	enc, err := crypto.Encrypt([]byte(n.Title), encryptionKey)
	if err != nil {
		log.Println(colors.RedText("Ошибка шифрования заголовка:"), err)
		return
	}
	n.Title = hex.EncodeToString(enc)
}

func (n *Note) DecryptTitle() string {
	if len(n.Title) < 32 || !isHex(n.Title) {
		return n.Title
	}
	encBytes, err := hex.DecodeString(n.Title)
	if err != nil {
		return n.Title
	}
	dec, err := crypto.Decrypt(encBytes, encryptionKey)
	if err != nil {
		return n.Title
	}
	return string(dec)
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}