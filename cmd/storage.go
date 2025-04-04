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

// saveNotesToMarkdown сохраняет все заметки в защищенный ZIP-архив
func saveNotesToMarkdown() error {
	if len(notes) == 0 {
		return nil // Не сохраняем пустые заметки
	}

	// Создаем буфер для Markdown содержимого
	var buf bytes.Buffer

	// Записываем данные в буфер
	_, err := buf.WriteString("# Менеджер паролей\n\n")
	if err != nil {
		return fmt.Errorf("ошибка при записи в буфер: %v", err)
	}

	_, err = buf.WriteString(fmt.Sprintf("> Последнее обновление: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		return fmt.Errorf("ошибка при записи в буфер: %v", err)
	}

	for _, note := range notes {
		decryptedPass := note.DecryptPassword()
		noteContent := fmt.Sprintf("## %s (ID: %d)\n", note.Title, note.ID)
		noteContent += fmt.Sprintf("- **Логин:** `%s`\n", note.Login)
		noteContent += fmt.Sprintf("- **Пароль:** `%s`\n", decryptedPass)
		noteContent += "\n---\n\n"
		_, err = buf.WriteString(noteContent)
		if err != nil {
			return fmt.Errorf("ошибка при записи заметки в буфер: %v", err)
		}
	}

	// Создаем защищенный ZIP-архив
	zipFile, err := os.Create(zipArchive)
	if err != nil {
		return fmt.Errorf("ошибка при создании архива: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("ARCHIVE_PASSWORD не найден в .env")
	}

	header := &zip.FileHeader{
		Name:   filepath.Base(storageFile),
		Method: zip.Deflate,
	}

	writer, err := zipWriter.Encrypt(header.Name, archivePassword, zip.AES256Encryption)
	if err != nil {
		return fmt.Errorf("ошибка при установке пароля на архив: %v", err)
	}

	_, err = io.Copy(writer, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("ошибка при записи в архив: %v", err)
	}

	return nil
}

// loadNotesFromMarkdown загружает заметки из защищенного ZIP-архива
func loadNotesFromMarkdown() error {
	// Проверяем, существует ли архив
	if _, err := os.Stat(zipArchive); os.IsNotExist(err) {
		return nil // Архив не существует, это нормально
	}

	archivePassword := os.Getenv("ARCHIVE_PASSWORD")
	if archivePassword == "" {
		return fmt.Errorf("ARCHIVE_PASSWORD не найден в .env")
	}

	// Открываем ZIP-архив
	r, err := zip.OpenReader(zipArchive)
	if err != nil {
		return fmt.Errorf("ошибка при открытии архива: %v", err)
	}
	defer r.Close()

	// Ищем наш файл в архиве
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

	// Устанавливаем пароль для файла
	file.SetPassword(archivePassword)

	// Открываем файл в архиве
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("ошибка при открытии файла в архиве: %v", err)
	}
	defer rc.Close()

	// Читаем содержимое файла
	scanner := bufio.NewScanner(rc)
	var currentNote *Note
	notes = []Note{} // Очищаем текущие заметки

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			if currentNote != nil {
				notes = append(notes, *currentNote)
			}
			// Парсим заголовок и ID: "## Title (ID: id)"
			start := len("## ")
			end := strings.Index(line, " (ID: ")
			if end == -1 {
				continue
			}
			title := line[start:end]
			idStr := line[end+len(" (ID: ") : len(line)-1] // Убираем ")"
			id, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			currentNote = &Note{ID: id, Title: title}
		} else if strings.HasPrefix(line, "- **Логин:** `") && currentNote != nil {
			start := len("- **Логин:** `")
			end := strings.LastIndex(line, "`")
			if end == -1 {
				continue
			}
			currentNote.Login = line[start:end]
		} else if strings.HasPrefix(line, "- **Пароль:** `") && currentNote != nil {
			start := len("- **Пароль:** `")
			end := strings.LastIndex(line, "`")
			if end == -1 {
				continue
			}
			plainPass := line[start:end]
			currentNote.EncryptPassword(plainPass)
		} else if line == "---" && currentNote != nil {
			notes = append(notes, *currentNote)
			currentNote = nil
		}
	}

	if currentNote != nil {
		notes = append(notes, *currentNote)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка при чтении файла: %v", err)
	}
	return nil
}
