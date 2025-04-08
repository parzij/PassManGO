package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unicode"

	"golang.org/x/term"
)

// readPasswordWithStars считывает пароль, маскируя символы звёздочками
func readPasswordWithStars() (string, error) {
	oldState, err := term.MakeRaw(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(syscall.Stdin), oldState)

	var password []rune
	reader := bufio.NewReader(os.Stdin)

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}
		if char == '\n' || char == '\r' {
			fmt.Println()
			break
		}
		// Обработка backspace
		if char == 127 || char == '\b' {
			if len(password) > 0 {
				password = password[:len(password)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		if unicode.IsPrint(char) {
			password = append(password, char)
			// Печатаем звёздочку
			fmt.Print(greenText("*"))
		}
	}
	return string(password), nil
}
