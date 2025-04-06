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

// saveNotesToMarkdown сохраняет все заметки в ZIP‑архив.
// Если архив уже существует, он перезаписывается на месте (без удаления файла).
func saveNotesToMarkdown() error {
	if len(notes) == 0 {
		return nil // Нет данных ‒ ничего не сохраняем
	}

	// Формируем Markdown‑содержимое
	var buf bytes.Buffer
	buf.WriteString("# Менеджер паролей\n\n")
	buf.WriteString(fmt.Sprintf("> Последнее обновление: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, note := range notes {
		buf.WriteString(fmt.Sprintf("## %s (ID: %d)\n", note.Title, note.ID))
		buf.WriteString(fmt.Sprintf("- **Логин:** `%s`\n", note.Login))
		buf.WriteString(fmt.Sprintf("- **Пароль:** `%s`\n", note.DecryptPassword()))
		buf.WriteString("\n---\n\n")
	}

	// Открываем файл архива (создаём, если нет) с перезаписью содержимого
	f, err := os.OpenFile(zipArchive, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("ошибка при открытии/создании архива: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("ARCHIVE_PASSWORD не найден в .env")
	}

	w, err := zw.Encrypt(filepath.Base(storageFile), archivePassword, zip.AES256Encryption)
	if err != nil {
		return fmt.Errorf("ошибка при установке пароля на архив: %v", err)
	}

	if _, err = io.Copy(w, bytes.NewReader(buf.Bytes())); err != nil {
		return fmt.Errorf("ошибка при записи в архив: %v", err)
	}

	return nil
}

// loadNotesFromMarkdown загружает заметки из архива (если он существует)
func loadNotesFromMarkdown() error {
	if _, err := os.Stat(zipArchive); os.IsNotExist(err) {
		return nil // Архива ещё нет ‒ это нормально
	}

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("ARCHIVE_PASSWORD не найден в .env")
	}

	r, err := zip.OpenReader(zipArchive)
	if err != nil {
		return fmt.Errorf("ошибка при открытии архива: %v", err)
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
		return fmt.Errorf("файл %s не найден в архиве", storageFile)
	}

	file.SetPassword(archivePassword)
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("ошибка при открытии файла в архиве: %v", err)
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	var current *Note
	notes = []Note{}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "## "):
			if current != nil {
				notes = append(notes, *current)
			}
			start := len("## ")
			end := strings.Index(line, " (ID: ")
			if end == -1 {
				continue
			}
			title := line[start:end]
			idStr := line[end+len(" (ID: ") : len(line)-1]
			id, _ := strconv.Atoi(idStr)
			current = &Note{ID: id, Title: title}
		case strings.HasPrefix(line, "- **Логин:** `") && current != nil:
			current.Login = strings.TrimSuffix(strings.TrimPrefix(line, "- **Логин:** `"), "`")
		case strings.HasPrefix(line, "- **Пароль:** `") && current != nil:
			pass := strings.TrimSuffix(strings.TrimPrefix(line, "- **Пароль:** `"), "`")
			current.EncryptPassword(pass)
		case line == "---" && current != nil:
			notes = append(notes, *current)
			current = nil
		}
	}
	if current != nil {
		notes = append(notes, *current)
	}

	return scanner.Err()
}
