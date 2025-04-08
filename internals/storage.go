package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yeka/zip"
)

// saveNotesToMarkdown сохраняет заметки (и избранные) в зашифрованный zip-архив
// (пароль совпадает с APP_PASSWORD)
func saveNotesToMarkdown() error {
	var buf bytes.Buffer
	buf.WriteString("# Менеджер паролей\n\n")
	buf.WriteString(fmt.Sprintf("> Последнее обновление: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Сохраняем обычные заметки
	for _, note := range notes {
		buf.WriteString(fmt.Sprintf("## %s (ID: %d)\n", note.Title, note.ID))
		buf.WriteString(fmt.Sprintf("- **Логин:** `%s`\n", note.Login))
		buf.WriteString(fmt.Sprintf("- **Пароль:** `%s`\n", note.DecryptPassword()))
		buf.WriteString(fmt.Sprintf("- **Избранное:** `%t`\n", note.Favorite))
		buf.WriteString("\n---\n\n")
	}

	// Сохраняем избранные отдельно
	buf.WriteString("# Избранные заметки ⭐\n\n")
	for _, note := range favorites {
		buf.WriteString(fmt.Sprintf("## %s (ID: %d)\n", note.Title, note.ID))
		buf.WriteString(fmt.Sprintf("- **Логин:** `%s`\n", note.Login))
		buf.WriteString(fmt.Sprintf("- **Пароль:** `%s`\n", note.DecryptPassword()))
		buf.WriteString("\n---\n\n")
	}

	f, err := os.OpenFile(zipArchive, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%sошибка при открытии/создании архива: %v%s", colorRed, err, colorReset)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("%sARCHIVE_PASSWORD не найден в .env%s", colorRed, colorReset)
	}

	w, err := zw.Encrypt(filepath.Base(storageFile), archivePassword, zip.AES256Encryption)
	if err != nil {
		return fmt.Errorf("%sошибка при установке пароля на архив: %v%s", colorRed, err, colorReset)
	}

	if _, err = io.Copy(w, bytes.NewReader(buf.Bytes())); err != nil {
		return fmt.Errorf("%sошибка при записи в архив: %v%s", colorRed, err, colorReset)
	}

	return nil
}

// loadNotesFromMarkdown загружает заметки (и избранные) из зашифрованного zip-архива
func loadNotesFromMarkdown() error {
	if _, err := os.Stat(zipArchive); os.IsNotExist(err) {
		return nil
	}

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("%sARCHIVE_PASSWORD не найден в .env%s", colorRed, colorReset)
	}

	r, err := zip.OpenReader(zipArchive)
	if err != nil {
		return fmt.Errorf("%sошибка при открытии архива: %v%s", colorRed, err, colorReset)
	}
	defer r.Close()

	var file *zip.File
	for _, f := range r.File {
		if f.Name == filepath.Base(storageFile) {
			file = f
			break
		}
	}
	if file == nil {
		return fmt.Errorf("%sфайл %s не найден в архиве%s", colorRed, storageFile, colorReset)
	}

	file.SetPassword(archivePassword)
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("%sошибка при открытии файла в архиве: %v%s", colorRed, err, colorReset)
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	var current *Note
	notes = []Note{}
	favorites = []Note{}
	inFavoritesSection := false

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "## "):
			// Сохраняем предыдущую заметку, если она была
			if current != nil {
				if inFavoritesSection {
					favorites = append(favorites, *current)
				} else {
					notes = append(notes, *current)
				}
			}
			start := len("## ")
			end := strings.Index(line, " (ID: ")
			if end == -1 {
				continue
			}
			title := line[start:end]
			idStr := line[end+len(" (ID: ") : len(line)-1]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			current = &Note{ID: id, Title: title}

		case strings.HasPrefix(line, "- **Логин:** `") && current != nil:
			current.Login = strings.TrimSuffix(strings.TrimPrefix(line, "- **Логин:** `"), "`")

		case strings.HasPrefix(line, "- **Пароль:** `") && current != nil:
			pass := strings.TrimSuffix(strings.TrimPrefix(line, "- **Пароль:** `"), "`")
			current.EncryptPassword(pass)

		case strings.HasPrefix(line, "- **Избранное:** `") && current != nil:
			fav, err := strconv.ParseBool(strings.TrimSuffix(strings.TrimPrefix(line, "- **Избранное:** `"), "`"))
			if err == nil {
				current.Favorite = fav
			}

		case strings.Contains(line, "# Избранные заметки ⭐"):
			inFavoritesSection = true

		case line == "---" && current != nil:
			// Добавляем текущую заметку в нужный список
			if inFavoritesSection {
				favorites = append(favorites, *current)
			} else {
				notes = append(notes, *current)
			}
			current = nil
		}
	}
	// Добавляем последнюю считанную заметку
	if current != nil {
		if inFavoritesSection {
			favorites = append(favorites, *current)
		} else {
			notes = append(notes, *current)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%sошибка при сканировании файла: %v%s", colorRed, err, colorReset)
	}

	return nil
}
