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
)

var (
	notes         []Note
	favorites     []Note
	storageFile   = "passwords.md"
	zipArchive    = "passwords.zip"
	encryptionKey []byte

	failedAttempts int       // счётчик неудачных попыток ввода пароля
	lastAttempt    time.Time // время последней неудачной попытки

	// Добавляем глобальные переменные для таймеров
	shutdownTimer *time.Timer
	warningTimer  *time.Timer
)

func main() {
	// Инициализация таймеров автоматического закрытия
	initShutdownTimers()

	// При выходе останавливаем таймеры
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
			// Проверяем, не заблокирована ли система
			if checkBlocked() {
				continue
			}

			fmt.Print(greenText("Введите пароль 🔑: "))
			input, err := readPasswordWithStars()
			if err != nil {
				log.Println(redText("Ошибка чтения пароля:"), err)
				continue
			}
			if input == appPassword {
				failedAttempts = 0 // сброс счетчика неудачных попыток
				break
			}
			failedAttempts++
			lastAttempt = time.Now()
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
		fmt.Println(redText("5) Выход 🚪"))
		fmt.Println(greenText("-----------------------------"))

		fmt.Print(greenText("Выберите действие: "))
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			addNote(reader)
		case "2":
			// Вызов под-меню для просмотра заметок (все/избранные)
			viewNotesMenu(reader)
		case "3":
			editCredentials(reader)
		case "4":
			deleteNote(reader)
		case "5":
			fmt.Println(redText("Выход из программы 👋"))
			if err := saveNotesToMarkdown(); err != nil {
				log.Println(redText("Ошибка при сохранении заметок:"), err)
			}
			return
		default:
			fmt.Println(redText("Неверный выбор, повторите ввод."))
		}
	}
}

// initShutdownTimers инициализирует таймеры для автоматического выхода
func initShutdownTimers() {
	// Таймер полного завершения (через 10 минут)
	shutdownTimer = time.AfterFunc(10*time.Minute, func() {
		fmt.Println(redText("Программа закрывается автоматически из-за бездействия."))
		if err := saveNotesToMarkdown(); err != nil {
			log.Println(redText("Ошибка при сохранении заметок:"), err)
		}
		os.Exit(0)
	})

	// Таймер предупреждения (за 1 минуту до завершения)
	warningTimer = time.AfterFunc(9*time.Minute, func() {
		fmt.Println(yellowText("Предупреждение: через 1 минуту программа будет закрыта за бездействие!"))
	})
}

// resetShutdownTimer сбрасывает оба таймера при каждом действии пользователя
func resetShutdownTimer() {
	if shutdownTimer != nil {
		shutdownTimer.Stop()
	}
	if warningTimer != nil {
		warningTimer.Stop()
	}
	// Запускаем их заново
	initShutdownTimers()
}

// checkBlocked проверяет, не заблокирована ли программа из-за превышения количества неудачных попыток
func checkBlocked() bool {
	if failedAttempts < 5 {
		return false
	}

	var blockDuration time.Duration
	switch {
	case failedAttempts >= 8:
		blockDuration = 5 * time.Minute
	case failedAttempts >= 6:
		blockDuration = 1 * time.Minute
	default:
		blockDuration = 30 * time.Second
	}

	elapsed := time.Since(lastAttempt)

	// Если ещё не истекло время блокировки
	if elapsed < blockDuration {
		remaining := blockDuration - elapsed
		fmt.Printf(yellowText("Система заблокирована на %v.\n"), blockDuration.Round(time.Second))
		fmt.Printf(yellowText("Повторите ввод через %v.\n"), remaining.Round(time.Second))
		time.Sleep(remaining)
		return true
	}

	// Если блокировка истекла, сбрасываем счётчик после 5 минут (при попытке снова ввести пароль)
	if failedAttempts >= 8 && elapsed >= 5*time.Minute {
		failedAttempts = 0
	}
	return false
}

// ensureConfig проверяет наличие пароля и ключа шифрования и, при необходимости, создаёт их
func ensureConfig() (bool, string) {
	_ = godotenv.Load(".env")

	appPass := os.Getenv("APP_PASSWORD")
	encKey := os.Getenv("ENCRYPTION_KEY")

	// Если в .env ещё нет пароля или ключа — просим пользователя ввести новый пароль
	if appPass == "" || encKey == "" {
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
		return true, pw
	}

	// Проверяем корректность ключа шифрования (должен быть длиной 32 байта = 64 hex-символа)
	if len(encKey) != 32 && len(encKey) != 64 {
		log.Fatalf(redText("ENCRYPTION_KEY должен быть 32 байта (hex 64 символа) ‒ у вас %d"), len(encKey))
	}

	// Сохраняем ключ шифрования в глобальную переменную
	keyBytes, err := hex.DecodeString(os.Getenv("ENCRYPTION_KEY"))
	if err != nil || len(keyBytes) != 32 {
		log.Fatal(redText("Некорректный ENCRYPTION_KEY в .env"))
	}
	encryptionKey = keyBytes

	return false, appPass
}
