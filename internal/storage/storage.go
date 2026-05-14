package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"parzij/PassManGO/internal/colors"
	"parzij/PassManGO/internal/note"

	"github.com/yeka/zip"
)

var (
	Notes     []note.Note
	Favorites []note.Note
)

const (
	storageFile = "passwords.md"
	zipArchive  = "passwords.zip"
)

// SaveNotesToMarkdown — сохраняет заметки в зашифрованный zip
func SaveNotesToMarkdown() error {
	var buf bytes.Buffer
	buf.WriteString("# Менеджер паролей\n\n")
	buf.WriteString(fmt.Sprintf("> Последнее обновление: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Обычные заметки
	for _, n := range Notes {
		decryptedTitle := n.DecryptTitle()
		buf.WriteString(fmt.Sprintf("## %s (ID: %d)\n", decryptedTitle, n.ID))
		buf.WriteString(fmt.Sprintf("- **Логин:** `%s`\n", n.Login))
		buf.WriteString(fmt.Sprintf("- **Пароль:** `%s`\n", n.DecryptPassword()))
		buf.WriteString(fmt.Sprintf("- **Избранное:** `%t`\n", n.Favorite))
		buf.WriteString("\n---\n\n")
	}

	// Избранные заметки
	if len(Favorites) > 0 {
		buf.WriteString("# Избранные заметки ⭐\n\n")
		for _, n := range Favorites {
			decryptedTitle := n.DecryptTitle()
			buf.WriteString(fmt.Sprintf("## %s (ID: %d)\n", decryptedTitle, n.ID))
			buf.WriteString(fmt.Sprintf("- **Логин:** `%s`\n", n.Login))
			buf.WriteString(fmt.Sprintf("- **Пароль:** `%s`\n", n.DecryptPassword()))
			buf.WriteString("\n---\n\n")
		}
	}

	f, err := os.OpenFile(zipArchive, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%sошибка создания архива: %v%s", colors.RedText(""), err, colors.GreenText(""))
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	archivePass := os.Getenv("ARCHIVE_PASSWORD")
	if archivePass == "" {
		return fmt.Errorf("%sARCHIVE_PASSWORD не найден в .env%s", colors.RedText(""), colors.GreenText(""))
	}

	w, err := zw.Encrypt(storageFile, archivePass, zip.AES256Encryption)
	if err != nil {
		return fmt.Errorf("ошибка шифрования архива: %v", err)
	}

	_, err = io.Copy(w, bytes.NewReader(buf.Bytes()))
	return err
}

// LoadNotesFromMarkdown — полный парсер (адаптирован)
func LoadNotesFromMarkdown() error {
	if _, err := os.Stat(zipArchive); os.IsNotExist(err) {
		Notes = []note.Note{}
		Favorites = []note.Note{}
		return nil
	}

	archivePass := os.Getenv("ARCHIVE_PASSWORD")
	if archivePass == "" {
		return fmt.Errorf("%sARCHIVE_PASSWORD не найден в .env%s", colors.RedText(""), colors.GreenText(""))
	}

	r, err := zip.OpenReader(zipArchive)
	if err != nil {
		return fmt.Errorf("ошибка открытия zip: %v", err)
	}
	defer r.Close()

	var zfile *zip.File
	for _, f := range r.File {
		if f.Name == storageFile {
			zfile = f
			break
		}
	}
	if zfile == nil {
		return fmt.Errorf("файл %s не найден в архиве", storageFile)
	}

	zfile.SetPassword(archivePass)
	rc, err := zfile.Open()
	if err != nil {
		return fmt.Errorf("ошибка открытия файла внутри архива: %v", err)
	}
	defer rc.Close()

	// ==================== ПАРСЕР ====================
	Notes = []note.Note{}
	Favorites = []note.Note{}
	var current *note.Note
	inFavoritesSection := false
	scanner := bufio.NewScanner(rc)

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "## "):
			if current != nil {
				if inFavoritesSection {
					Favorites = append(Favorites, *current)
				} else {
					Notes = append(Notes, *current)
				}
			}

			start := len("## ")
			end := strings.Index(line, " (ID: ")
			if end == -1 {
				continue
			}

			title := line[start:end]
			idStr := line[end+len(" (ID: "): len(line)-1]
			id, _ := strconv.Atoi(idStr)

			current = &note.Note{ID: id, Title: title}

		case strings.HasPrefix(line, "- **Логин:** `") && current != nil:
			current.Login = strings.TrimSuffix(strings.TrimPrefix(line, "- **Логин:** `"), "`")

		case strings.HasPrefix(line, "- **Пароль:** `") && current != nil:
			pass := strings.TrimSuffix(strings.TrimPrefix(line, "- **Пароль:** `"), "`")
			current.EncryptPassword(pass)

		case strings.HasPrefix(line, "- **Избранное:** `") && current != nil:
			favStr := strings.TrimSuffix(strings.TrimPrefix(line, "- **Избранное:** `"), "`")
			current.Favorite, _ = strconv.ParseBool(favStr)

		case strings.Contains(line, "# Избранные заметки ⭐"):
			inFavoritesSection = true

		case line == "---" && current != nil:
			if inFavoritesSection {
				Favorites = append(Favorites, *current)
			} else {
				Notes = append(Notes, *current)
			}
			current = nil
		}
	}

	// Последняя заметка
	if current != nil {
		if inFavoritesSection {
			Favorites = append(Favorites, *current)
		} else {
			Notes = append(Notes, *current)
		}
	}

	return scanner.Err()
}

// ====================== Геттеры ======================
func GetNotes() []note.Note {
	return Notes
}

func GetFavorites() []note.Note {
	return Favorites
}