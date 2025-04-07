package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	notes         []Note
	storageFile   = "passwords.md"
	zipArchive    = "passwords.zip"
	encryptionKey []byte
)

func main() {
	// Инициализация таймеров автоматического закрытия
	initShutdownTimers()
	defer func() {
		if shutdownTimer != nil {
			shutdownTimer.Stop()
		}
		if warningTimer != nil {
			warningTimer.Stop()
		}
	}()

	// Проверяем конфигурацию и, при необходимости, создаём её
	firstRun, appPassword := ensureConfig()

	// Если это не первый запуск - просим пользователя ввести пароль
	if !firstRun {
		for {
			fmt.Print(greenText("Введите пароль 🔑: "))
			input, err := readPasswordWithStars()
			if err != nil {
				log.Println(redText("Ошибка чтения пароля:"), err)
				continue
			}
			if input == appPassword {
				break
			}
			fmt.Println(redText("Неверный пароль. Попробуйте ещё раз.\n"))
		}
	}

	// Автоматическая загрузка заметок при старте
	if err := loadNotesFromMarkdown(); err != nil {
		log.Println(redText("Ошибка при загрузке заметок:"), err)
	} else if len(notes) > 0 {
		fmt.Printf(greenText("Загружено %d заметок 📂\n"), len(notes))
	} else {
		fmt.Println(greenText("Заметок не найдено. Вы можете добавить новую заметку ✨"))
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		resetShutdownTimer() // Сбрасываем таймер при каждом действии пользователя

		fmt.Println(greenText("\nГлавное меню 📋"))
		fmt.Println(greenText("-----------------------------"))
		fmt.Println(greenText("1) Добавить заметку ➕"))
		fmt.Println(greenText("2) Посмотреть заметки 👀"))
		fmt.Println(greenText("3) Редактировать заметку ✏️"))
		fmt.Println(greenText("4) Удалить заметку ❌"))
		fmt.Println(redText("5) Выход 🚪")) // Пункт выхода красным цветом
		fmt.Println(greenText("-----------------------------"))

		fmt.Print(greenText("Выберите действие: "))
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
			fmt.Println(redText("Выход из программы 👋")) // Сообщение о выходе красным
			if err := saveNotesToMarkdown(); err != nil {
				log.Println(redText("Ошибка при сохранении заметок:"), err)
			}
			return
		default:
			fmt.Println(redText("Неверный выбор, повторите ввод."))
		}
	}
}

func ensureConfig() (firstRun bool, appPass string) {
	_ = godotenv.Load(".env")

	appPass = os.Getenv("APP_PASSWORD")
	encKey := os.Getenv("ENCRYPTION_KEY")

	if appPass == "" || encKey == "" {
		firstRun = true
		fmt.Println(greenText("👋 Похоже, это первый запуск программы!"))
		fmt.Print(greenText("Введите новый пароль для приложения 🔐: "))
		pw, err := readPasswordWithStars()
		if err != nil || strings.TrimSpace(pw) == "" {
			log.Fatal(redText("Пароль не может быть пустым."))
		}

		// Генерируем случайный 32-байтный ключ шифрования
		randomKey := make([]byte, 32)
		if _, err := rand.Read(randomKey); err != nil {
			log.Fatal(redText("Не удалось сгенерировать ключ шифрования:"), err)
		}
		encKeyHex := hex.EncodeToString(randomKey)

		// Создаём/перезаписываем .env
		envContent := fmt.Sprintf("APP_PASSWORD=%s\nARCHIVE_PASSWORD=%s\nENCRYPTION_KEY=%s\n",
			pw, pw, encKeyHex)
		if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
			log.Fatal(redText("Не удалось записать .env:"), err)
		}

		// Обновляем переменные окружения текущего процесса
		_ = os.Setenv("APP_PASSWORD", pw)
		_ = os.Setenv("ARCHIVE_PASSWORD", pw)
		_ = os.Setenv("ENCRYPTION_KEY", encKeyHex)

		fmt.Println(greenText("✅ Пароль успешно установлен! Запустите программу снова, чтобы начать работу."))
		appPass = pw
	} else {
		firstRun = false
	}

	// Проверяем корректность ключа шифрования
	if encKey := os.Getenv("ENCRYPTION_KEY"); len(encKey) != 32 && len(encKey) != 64 {
		log.Fatalf(redText("ENCRYPTION_KEY должен быть 32 байта (hex 64 символа) ‒ у вас %d"), len(encKey))
	}

	// Сохраняем ключ шифрования в глобальную переменную
	keyBytes, err := hex.DecodeString(os.Getenv("ENCRYPTION_KEY"))
	if err != nil || len(keyBytes) != 32 {
		log.Fatal(redText("Некорректный ENCRYPTION_KEY в .env"))
	}
	encryptionKey = keyBytes

	return
}
