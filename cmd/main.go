package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	notes         []Note
	storageFile   = "passwords.md"  // имя файла внутри архива
	zipArchive    = "passwords.zip" // имя архива
	encryptionKey []byte
)

func main() {
	// Загрузка конфигурации из .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла:", err)
	}

	// Проверка пароля приложения
	correctPassword := os.Getenv("APP_PASSWORD")
	if correctPassword == "" {
		log.Fatal("APP_PASSWORD не найден в .env")
	}

	// Проверка ключа шифрования
	encKey := os.Getenv("ENCRYPTION_KEY")
	if len(encKey) == 0 {
		log.Fatal("ENCRYPTION_KEY не найден в .env")
	}
	if len(encKey) != 32 {
		log.Fatalf("ENCRYPTION_KEY должен быть 32 байта, а у вас %d", len(encKey))
	}
	encryptionKey = []byte(encKey)

	// Аутентификация пользователя
	for {
		fmt.Print("Введите пароль: ")
		input, err := readPasswordWithStars()
		if err != nil {
			log.Println("Ошибка чтения пароля:", err)
			continue
		}
		if input == correctPassword {
			break
		}
		fmt.Println("Неверный пароль. Попробуйте ещё раз.\n")
	}

	// Автоматическая загрузка заметок при старте
	if err := loadNotesFromMarkdown(); err != nil {
		log.Println("Ошибка при загрузке заметок:", err)
	} else if len(notes) > 0 {
		fmt.Printf("Загружено %d заметок\n", len(notes))
	} else {
		fmt.Println("Заметок не найдено. Вы можете добавить новую заметку.")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nГлавное меню 📋")
		fmt.Println("-----------------------------")
		fmt.Println("1) Добавить заметку ➕")
		fmt.Println("2) Посмотреть заметки 👀")
		fmt.Println("3) Редактировать заметку ✏️")
		fmt.Println("4) Удалить заметку ❌")
		fmt.Println("5) Выход 🚪")
		fmt.Println("-----------------------------")

		fmt.Print("Выберите действие: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			addNote(reader)
		case "2":
			viewNotes()
		case "3":
			editCredentials(reader)
		case "4":
			deleteNote(reader)
		case "5":
			fmt.Println("Выход из программы.")
			if err := saveNotesToMarkdown(); err != nil {
				log.Println("Ошибка при сохранении заметок:", err)
			}
			return
		default:
			fmt.Println("Неверный выбор, повторите ввод.")
		}
	}
}
