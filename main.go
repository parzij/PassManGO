package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"parzij/PassManGO/internal/colors"
	"parzij/PassManGO/internal/note"
	"parzij/PassManGO/internal/operations"
	"parzij/PassManGO/internal/storage"
	"parzij/PassManGO/internal/tui"
)

var (
	encryptionKey []byte

	failedAttempts int
	lastAttempt    time.Time

	shutdownTimer *time.Timer
	warningTimer  *time.Timer
)

func main() {
	initShutdownTimers()

	defer func() {
		if shutdownTimer != nil {
			shutdownTimer.Stop()
		}
		if warningTimer != nil {
			warningTimer.Stop()
		}
	}()

	firstRun, _ := ensureConfig()

	if !firstRun {
		note.SetEncryptionKey(encryptionKey)

		if err := storage.LoadNotesFromMarkdown(); err != nil {
			log.Println(colors.RedText("Ошибка при загрузке заметок:"), err)
		} else if len(storage.Notes) > 0 {
			fmt.Printf(colors.GreenText("Загружено %d заметок 📂\n"), len(storage.Notes))
		} else {
			fmt.Println(colors.GreenText("Заметок не найдено. Добавьте первую ✨"))
		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		ResetShutdownTimer()

		fmt.Println(colors.GreenText("\nГлавное меню 📋"))
		fmt.Println(colors.GreenText("-----------------------------"))
		fmt.Println(colors.GreenText("1) Добавить заметку ➕"))
		fmt.Println(colors.GreenText("2) Посмотреть заметки 👀"))
		fmt.Println(colors.GreenText("3) Редактировать заметку ✏️"))
		fmt.Println(colors.GreenText("4) Удалить заметку ❌"))
		fmt.Println(colors.RedText("5) Выход 🚪"))
		fmt.Println(colors.GreenText("-----------------------------"))

		fmt.Print(colors.GreenText("Выберите действие: "))
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			operations.AddNote(reader)
		case "2":
			operations.ViewNotesMenu(reader)
		case "3":
			operations.EditCredentials(reader)
		case "4":
			operations.DeleteNote(reader)
		case "5":
			fmt.Println(colors.RedText("Выход из программы 👋"))
			if err := storage.SaveNotesToMarkdown(); err != nil {
				log.Println(colors.RedText("Ошибка сохранения:"), err)
			}
			return
		default:
			fmt.Println(colors.RedText("Неверный выбор, повторите ввод."))
		}
	}
}

// ====================== ensureConfig ======================
func ensureConfig() (bool, string) {
	_ = godotenv.Load(".env")

	appPass := os.Getenv("APP_PASSWORD")
	encKey := os.Getenv("ENCRYPTION_KEY")

	if appPass == "" || encKey == "" {
		fmt.Println(colors.GreenText("👋 Первый запуск программы!"))
		fmt.Print(colors.GreenText("Введите новый мастер-пароль 🔐: "))

		pw, err := tui.ReadPasswordWithStars()
		if err != nil || strings.TrimSpace(pw) == "" {
			log.Fatal(colors.RedText("Пароль не может быть пустым."))
		}

		randomKey := make([]byte, 32)
		if _, err := rand.Read(randomKey); err != nil {
			log.Fatal(colors.RedText("Не удалось сгенерировать ключ"))
		}

		encKeyHex := hex.EncodeToString(randomKey)

		envContent := fmt.Sprintf("APP_PASSWORD=%s\nARCHIVE_PASSWORD=%s\nENCRYPTION_KEY=%s\n", pw, pw, encKeyHex)

		if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
			log.Fatal(colors.RedText("Не удалось создать .env"))
		}

		fmt.Println(colors.GreenText("✅ Пароль установлен! Перезапустите программу."))
		os.Exit(0)
	}

	keyBytes, err := hex.DecodeString(encKey)
	if err != nil || len(keyBytes) != 32 {
		log.Fatal(colors.RedText("Некорректный ENCRYPTION_KEY (нужно 64 hex символа)"))
	}
	encryptionKey = keyBytes

	return false, appPass
}

// ====================== Таймеры ======================
func initShutdownTimers() {
	shutdownTimer = time.AfterFunc(3*time.Minute, func() {
		fmt.Println(colors.RedText("\nПрограмма закрывается из-за бездействия."))
		storage.SaveNotesToMarkdown()
		os.Exit(0)
	})

	warningTimer = time.AfterFunc(2*time.Minute, func() {
		fmt.Println(colors.YellowText("⚠️ Через 1 минуту программа закроется!"))
	})
}

func ResetShutdownTimer() {
	if shutdownTimer != nil {
		shutdownTimer.Stop()
	}
	if warningTimer != nil {
		warningTimer.Stop()
	}
	initShutdownTimers()
}