package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
)

// addNote добавляет новую заметку.
func addNote(reader *bufio.Reader) {
	fmt.Println("\n--- Добавление заметки ---")
	fmt.Println("0) Назад")
	fmt.Print("Введите заголовок или 0 - для отмены: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	if title == "0" {
		return
	}

	fmt.Print("Введите логин: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Введите пароль: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	note := Note{ID: len(notes) + 1, Title: title, Login: login}
	note.EncryptPassword(pass)
	notes = append(notes, note)
	fmt.Printf("Заметка с ID=%d успешно добавлена!\n", note.ID)

	if err := saveNotesToMarkdown(); err != nil {
		log.Println("Ошибка при сохранении заметок:", err)
	}
}

// editCredentials редактирует логин или пароль заметки.
func editCredentials(reader *bufio.Reader) {
	fmt.Println("\n--- Редактирование логина/пароля ---")
	if len(notes) == 0 {
		fmt.Println("Список заметок пуст.")
		return
	}

	fmt.Print("Введите ID заметки (или 0 для отмены): ")
	var id int
	_, err := fmt.Scanf("%d\n", &id)
	if err != nil {
		fmt.Println("Ошибка ввода ID:", err)
		return
	}
	if id == 0 {
		return
	}

	for i, note := range notes {
		if note.ID == id {
			fmt.Printf("Текущий логин: %s\n", note.Login)
			fmt.Printf("Текущий пароль: %s\n", note.DecryptPassword())
			fmt.Println("0) Назад")
			fmt.Println("1) Логин")
			fmt.Println("2) Пароль")
			fmt.Print("Введите номер выбора: ")
			selection, _ := reader.ReadString('\n')
			selection = strings.TrimSpace(selection)

			if selection == "0" {
				return
			}

			switch selection {
			case "1":
				fmt.Print("Введите новый логин: ")
				newLogin, _ := reader.ReadString('\n')
				notes[i].Login = strings.TrimSpace(newLogin)
				fmt.Println("Логин успешно изменён!")
			case "2":
				fmt.Print("Введите новый пароль: ")
				newPass, _ := reader.ReadString('\n')
				notes[i].EncryptPassword(strings.TrimSpace(newPass))
				fmt.Println("Пароль успешно изменён!")
			default:
				fmt.Println("Неверный выбор.")
				return
			}
			if err := saveNotesToMarkdown(); err != nil {
				log.Println("Ошибка при сохранении заметок:", err)
			}
			return
		}
	}
	fmt.Println("Заметка с таким ID не найдена.")
}

// deleteNote удаляет заметку по ID и перенумеровывает оставшиеся.
func deleteNote(reader *bufio.Reader) {
	fmt.Println("\n--- Удаление заметки ---")
	if len(notes) == 0 {
		fmt.Println("Список заметок пуст.")
		return
	}

	fmt.Print("Введите ID заметки или 0 - для отмены: ")
	var id int
	_, err := fmt.Scanf("%d\n", &id)
	if err != nil {
		fmt.Println("Ошибка ввода ID:", err)
		return
	}
	if id == 0 {
		return
	}

	for i, note := range notes {
		if note.ID == id {
			notes = append(notes[:i], notes[i+1:]...)
			reindexNotes()
			fmt.Printf("Заметка с ID=%d удалена.\n", id)
			if err := saveNotesToMarkdown(); err != nil {
				log.Println("Ошибка при сохранении заметок:", err)
			}
			return
		}
	}
	fmt.Println("Заметка с таким ID не найдена.")
}

// reindexNotes перенумеровывает ID заметок.
func reindexNotes() {
	for i := range notes {
		notes[i].ID = i + 1
	}
}

// viewNotes отображает заметки в виде таблицы.
func viewNotes() {
	fmt.Println("\n--- Список заметок ---")
	if len(notes) == 0 {
		fmt.Println("Список заметок пуст.")
		return
	}

	const tableWidth = 86
	fmt.Println(strings.Repeat("-", tableWidth))
	fmt.Printf("| %-3s | %-20s | %-30s | %-20s |\n", "№", "Заголовок", "Логин", "Пароль")
	fmt.Println(strings.Repeat("-", tableWidth))

	for _, note := range notes {
		fmt.Printf("| %-3d | %-20s | %-30s | %-20s |\n",
			note.ID, note.Title, note.Login, note.DecryptPassword())
	}
	fmt.Println(strings.Repeat("-", tableWidth))
}
