package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"unicode"
)

func addNote(reader *bufio.Reader) {
	fmt.Println(greenText("\n--- Добавление заметки ➕ ---"))
	fmt.Println(redText("0) Назад"))
	fmt.Print(greenText("Введите заголовок или 0 - для отмены: "))
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	if title == "0" {
		return
	}

	fmt.Print(greenText("Введите логин: "))
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print(greenText("Введите пароль: "))
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	strength := evaluatePasswordStrength(pass)
	fmt.Printf("%sОценка сложности пароля: %s%s\n", colorGreen, strength, colorReset)

	note := Note{ID: len(notes) + 1, Title: title, Login: login}
	note.EncryptPassword(pass)
	notes = append(notes, note)
	fmt.Printf(greenText("Заметка с ID=%d успешно добавлена!\n"), note.ID)

	if err := saveNotesToMarkdown(); err != nil {
		log.Println(redText("Ошибка при сохранении заметок:"), err)
	}
}

func evaluatePasswordStrength(password string) string {
	if len(password) == 0 {
		return redText("❌ Пустой пароль")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
		length     = len(password)
		score      int
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if length > 12 {
		score += 3
	} else if length > 8 {
		score += 2
	} else if length > 5 {
		score += 1
	}

	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	switch {
	case score >= 7:
		return greenText("🔒 Очень сильный")
	case score >= 5:
		return greenText("🔐 Сильный")
	case score >= 3:
		return yellowText("⚠️ Средний")
	default:
		return redText("❌ Слабый")
	}
}

func editCredentials(reader *bufio.Reader) {
	fmt.Println(greenText("\n--- Редактирование заметки ✏️ ---"))
	if len(notes) == 0 {
		fmt.Println(redText("Список заметок пуст."))
		return
	}

	fmt.Print(greenText("Введите ID заметки (или 0 для отмены): "))
	var id int
	_, err := fmt.Scanf("%d\n", &id)
	if err != nil {
		fmt.Println(redText("Ошибка ввода ID:"), err)
		return
	}
	if id == 0 {
		return
	}

	for i, note := range notes {
		if note.ID == id {
			fmt.Printf(greenText("Текущий логин: %s\n"), note.Login)
			fmt.Printf(greenText("Текущий пароль: %s\n"), note.DecryptPassword())
			fmt.Println(redText("0) Назад"))
			fmt.Println(greenText("1) Логин"))
			fmt.Println(greenText("2) Пароль"))
			fmt.Print(greenText("Введите номер выбора: "))
			selection, _ := reader.ReadString('\n')
			selection = strings.TrimSpace(selection)

			if selection == "0" {
				return
			}

			switch selection {
			case "1":
				fmt.Print(greenText("Введите новый логин: "))
				newLogin, _ := reader.ReadString('\n')
				notes[i].Login = strings.TrimSpace(newLogin)
				fmt.Println(greenText("Логин успешно изменён!"))
			case "2":
				fmt.Print(greenText("Введите новый пароль: "))
				newPass, _ := reader.ReadString('\n')
				notes[i].EncryptPassword(strings.TrimSpace(newPass))
				strength := evaluatePasswordStrength(notes[i].DecryptPassword())
				fmt.Printf("%sОценка сложности нового пароля: %s%s\n", colorGreen, strength, colorReset)
				fmt.Println(greenText("Пароль успешно изменён!"))
			default:
				fmt.Println(redText("Неверный выбор."))
				return
			}
			if err := saveNotesToMarkdown(); err != nil {
				log.Println(redText("Ошибка при сохранении заметок:"), err)
			}
			return
		}
	}
	fmt.Println(redText("Заметка с таким ID не найдена."))
}

func deleteNote(reader *bufio.Reader) {
	fmt.Println(greenText("\n--- Удаление заметки ❌ ---"))
	if len(notes) == 0 {
		fmt.Println(redText("Список заметок пуст."))
		return
	}

	fmt.Print(greenText("Введите ID заметки или 0 - для отмены: "))
	var id int
	_, err := fmt.Scanf("%d\n", &id)
	if err != nil {
		fmt.Println(redText("Ошибка ввода ID:"), err)
		return
	}
	if id == 0 {
		return
	}

	for i, note := range notes {
		if note.ID == id {
			notes = append(notes[:i], notes[i+1:]...)
			reindexNotes()
			fmt.Printf(greenText("Заметка с ID=%d удалена.\n"), id)
			if err := saveNotesToMarkdown(); err != nil {
				log.Println(redText("Ошибка при сохранении заметок:"), err)
			}
			return
		}
	}
	fmt.Println(redText("Заметка с таким ID не найдена."))
}

func reindexNotes() {
	for i := range notes {
		notes[i].ID = i + 1
	}
}

func viewNotes() {
	fmt.Println(greenText("\n--- Список заметок 👀 ---"))
	if len(notes) == 0 {
		fmt.Println(redText("Список заметок пуст."))
		return
	}

	const tableWidth = 86
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))
	fmt.Printf(greenText("| %-3s | %-20s | %-30s | %-20s |\n"), "№", "Заголовок", "Логин", "Пароль")
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))

	for _, note := range notes {
		fmt.Printf(greenText("| %-3d | %-20s | %-30s | %-20s |\n"),
			note.ID, note.Title, note.Login, note.DecryptPassword())
	}
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))
}
