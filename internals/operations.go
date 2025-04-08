package main

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

// addNote добавляет новую заметку
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
	fmt.Printf("%sОценка сложности пароля: %s\n", greenText(""), strength)

	note := Note{
		ID:    len(notes) + 1,
		Title: title,
		Login: login,
	}
	note.EncryptPassword(pass)
	note.EncryptTitle()

	notes = append(notes, note)
	fmt.Printf(greenText("Заметка с ID=%d успешно добавлена!\n"), note.ID)

	if err := saveNotesToMarkdown(); err != nil {
		log.Println(redText("Ошибка при сохранении заметок:"), err)
	}
}

// viewNotesMenu отображает подменю выбора: все заметки или избранные
func viewNotesMenu(reader *bufio.Reader) {
	for {
		fmt.Println(greenText("\n--- Просмотр заметок ---"))
		fmt.Println(redText("0) Назад"))
		fmt.Println(greenText("1) Избранные заметки"))
		fmt.Println(greenText("2) Все заметки"))
		fmt.Print(greenText("Выберите действие: "))

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "0":
			return
		case "1":
			// Просмотр избранных заметок
			viewNotes(reader, true)
		case "2":
			// Просмотр всех заметок
			viewNotes(reader, false)
		default:
			fmt.Println(redText("Неверный выбор, повторите ввод."))
		}
	}
}

// viewNotes выводит список заметок или избранных заметок в зависимости от флага showFavorites
func viewNotes(reader *bufio.Reader, showFavorites bool) {
	var displayNotes []Note
	title := "Список заметок 👀"

	if showFavorites {
		displayNotes = favorites
		title = "Избранные заметки ⭐"
	} else {
		displayNotes = notes
	}

	fmt.Println(greenText("\n--- " + title + " ---"))
	if len(displayNotes) == 0 {
		fmt.Println(redText("Список заметок пуст."))
		return
	}

	const tableWidth = 86
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))
	fmt.Printf(greenText("| %-3s | %-20s | %-30s | %-20s |\n"), "№", "Заголовок", "Логин", "Пароль")
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))

	for _, note := range displayNotes {
		decryptedTitle := note.DecryptTitle()
		if len(decryptedTitle) > 20 {
			decryptedTitle = decryptedTitle[:17] + "..."
		}
		fmt.Printf(greenText("| %-3d | %-20s | %-30s | %-20s |\n"),
			note.ID, decryptedTitle, note.Login, note.DecryptPassword())
	}
	fmt.Println(greenText(strings.Repeat("-", tableWidth)))

	if showFavorites {
		// Если просматриваем избранные заметки
		fmt.Print(greenText("\nВведите ID заметки, чтобы удалить её из избранного (0 - назад): "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" {
			return
		}

		id, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println(redText("Неверный ID заметки"))
			return
		}

		// Удаляем из избранного
		removeFromFavorites(id)
		fmt.Printf(greenText("Заметка с ID=%d удалена из избранного\n"), id)

		if err := saveNotesToMarkdown(); err != nil {
			log.Println(redText("Ошибка при сохранении заметок:"), err)
		}

	} else {
		// Если просматриваем общий список, даём возможность добавить/удалить из избранного
		fmt.Print(greenText("\nВведите ID заметки, чтобы добавить/удалить из избранного (0 - назад): "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" {
			return
		}

		id, err := strconv.Atoi(input)
		if err != nil || id < 1 || id > len(notes) {
			fmt.Println(redText("Неверный ID заметки"))
			return
		}

		for i, note := range notes {
			if note.ID == id {
				notes[i].Favorite = !notes[i].Favorite
				if notes[i].Favorite {
					favorites = append(favorites, notes[i])
					fmt.Printf(greenText("Заметка с ID=%d добавлена в избранное ⭐\n"), id)
				} else {
					removeFromFavorites(id)
					fmt.Printf(greenText("Заметка с ID=%d удалена из избранного\n"), id)
				}
				break
			}
		}

		if err := saveNotesToMarkdown(); err != nil {
			log.Println(redText("Ошибка при сохранении заметок:"), err)
		}
	}
}

// removeFromFavorites удаляет заметку из списка избранных
func removeFromFavorites(id int) {
	for i, note := range favorites {
		if note.ID == id {
			favorites = append(favorites[:i], favorites[i+1:]...)
			return
		}
	}
}

// editCredentials (пример функции — без изменений)
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

// deleteNote удаляет заметку
func deleteNote(reader *bufio.Reader) {
	fmt.Println(redText("\n--- Удаление заметки ❌ ---"))
	fmt.Println(redText("0) Назад"))
	fmt.Print(greenText("Введите ID заметки или 0 - для отмены: "))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "0" {
		return
	}

	id, err := strconv.Atoi(input)
	if err != nil || id < 1 || id > len(notes) {
		fmt.Println(redText("Неверный ID заметки"))
		return
	}

	for i, note := range notes {
		if note.ID == id {
			// Удаляем из notes
			notes = append(notes[:i], notes[i+1:]...)
			// Также нужно удалить из favorites, если оно там было
			removeFromFavorites(id)
			fmt.Printf(greenText("Заметка с ID=%d успешно удалена\n"), id)
			break
		}
	}

	if err := saveNotesToMarkdown(); err != nil {
		log.Println(redText("Ошибка при сохранении заметок:"), err)
	}
}

// evaluatePasswordStrength — пример оценки сложности пароля
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
