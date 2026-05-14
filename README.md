# 🗝️ Console Note Manager

### 🔐 Умный, безопасный и удобный способ управлять своими логинами и паролями прямо из терминала.

---

## 🏗️ Архитектура проекта

```
/PassManGO
├── main.go             # Точка входа, главное меню и конфигурация
├── go.mod              # Зависимости проекта
├── .env.example        # Конфигурация
├── passwords.zip       # Зашифрованное хранилище заметок
│
└── internal/           # Внутренние пакеты
├── colors/             # Цвета и форматирование вывода
│   └── colors.go
├── crypto/             # Шифрование AES
│   └── crypto.go
├── note/               # Структура Note
│   └── note.go
├── operations/         # Бизнес-логика
│   └── operations.go
├── storage/            # Сохранение и загрузка из zip-архива
│   └── storage.go
└── tui/                # Пользовательский интерфейс
└── ui.go
```
---

## 📌 Цель проекта

Создание безопасного консольного приложения на языке **Go**, позволяющего пользователю:
- Ввести **мастер-пароль** (настроен в `.env`).
- Добавлять, удалять, просматривать и редактировать заметки.
- Каждая заметка может содержать: **заголовок**, **логин**, **зашифрованный пароль**, **текст**.
- Интерфейс ввода пароля отображает `*` при вводе символов.
- Все чувствительные данные шифруются с использованием **AES-256**.

---

## ⚙️ Используемые технологии

| Технология | Назначение |
|-----------|------------|
| `Go` | Основной язык программирования |
| `AES-256` | Шифрование паролей |
| `bufio`, `os`, `fmt` | Работа с консолью и вводом |
| `syscall`, `golang.org/x/term` | Реализация маскированного ввода (`*`) |
| `github.com/joho/godotenv` | Загрузка конфигурации из `.env` |
| `hex` | Кодирование зашифрованных данных в текстовый формат |

---

## 💡 Основной функционал

### 1. Ввод мастер-пароля с маской

```go
// internal/tui/ui.go
func ReadPasswordWithStars() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var password []rune
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(colors.GreenText("Введите пароль 🔑: "))

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		if char == '\n' || char == '\r' {
			fmt.Println()
			break
		}

		if char == 127 || char == '\b' { // Backspace
			if len(password) > 0 {
				password = password[:len(password)-1]
				fmt.Print("\b \b")
			}
			continue
		}

		if unicode.IsPrint(char) {
			password = append(password, char)
			fmt.Print(colors.GreenText("*"))
		}
	}

	return string(password), nil
}
```

🔸 Назначение: Безопасный маскированный ввод пароля в терминале с поддержкой Backspace и цветным оформлением.

### 2. Шифрование и расшифровка данных
```go
// internal/crypto/crypto.go
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	// Генерируем случайный Initialization Vector
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
```

🔸 Назначение: Надёжное AES-CFB шифрование всех чувствительных данных (пароли и заголовки заметок).

### 3. Структура заметки
```go
// internal/note/note.go
type Note struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Login    string `json:"login"`
	Pass     string `json:"pass"`     // Зашифрованный пароль в hex
	Favorite bool   `json:"favorite"`
}
```
Основные методы:

- EncryptPassword() — шифрует пароль
- DecryptPassword() — расшифровывает пароль
- EncryptTitle() / DecryptTitle() — шифрование заголовка
- SetEncryptionKey() — установка глобального ключа



---

## 📋 Пример интерфейса в терминале

```
Главное меню 📋
-----------------------------
1) Добавить заметку ➕
2) Посмотреть заметки 👀
3) Редактировать заметку ✏️
4) Удалить заметку ❌
5) Выход 🚪
-----------------------------
Выберите действие:
```

---


## 📊 Пример таблицы заметок

```
---------------------------------------------------------------------------------
| №   | Заголовок           | Логин                        | Пароль             |
---------------------------------------------------------------------------------
| 1   | Google              | youremail@gmail.com          | yourpassword       |
| 2   | GitHub              | mygithub@gh.com              | mysecretpass       |
---------------------------------------------------------------------------------
```

---

## 🔒 Безопасность

- Все пароли хранятся в зашифрованном виде (AES-256).
- Мастер-пароль и ключ находятся только в `.env` файле, не загружаются в репозиторий.
- Ввод пароля всегда маскируется.
