package tui

import (
	"bufio"
	"fmt"
	"os"
	"unicode"

	"parzij/PassManGO/internal/colors"

	"golang.org/x/term"
)

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

		// Backspace
		if char == 127 || char == '\b' {
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

// ResetShutdownTimer — вызывается из main
func ResetShutdownTimer() {
	// Эта функция будет вызывать ResetShutdownTimer из main.go
	// Для этого сделаем небольшую обёртку (см. ниже)
}