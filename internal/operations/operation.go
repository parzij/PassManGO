package operations

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"parzij/PassManGO/internal/colors"
	"parzij/PassManGO/internal/note"
	"parzij/PassManGO/internal/storage"
)

func AddNote(reader *bufio.Reader) {
	fmt.Println(colors.GreenText("\n--- Добавление заметки ➕ ---"))
	fmt.Println(colors.RedText("0) Назад"))

	fmt.Print(colors.GreenText("Введите заголовок: "))
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	if title == "0" || title == "" {
		return
	}

	fmt.Print(colors.GreenText("Введите логин: "))
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print(colors.GreenText("Введите пароль: "))
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	n := note.Note{
		ID:    len(storage.Notes) + 1,
		Title: title,
		Login: login,
	}
	n.EncryptPassword(pass)
	n.EncryptTitle()

	storage.Notes = append(storage.Notes, n)

	fmt.Printf(colors.GreenText("Заметка ID=%d успешно добавлена!\n"), n.ID)

	if err := storage.SaveNotesToMarkdown(); err != nil {
		log.Println(colors.RedText("Ошибка сохранения:"), err)
	}
}

func ViewNotesMenu(reader *bufio.Reader) {
	for {
		fmt.Println(colors.GreenText("\n--- Просмотр заметок ---"))
		fmt.Println(colors.RedText("0) Назад"))
		fmt.Println(colors.GreenText("1) Избранные заметки"))
		fmt.Println(colors.GreenText("2) Все заметки"))
		fmt.Print(colors.GreenText("Выберите: "))

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "0":
			return
		case "1":
			viewNotes(true)
		case "2":
			viewNotes(false)
		default:
			fmt.Println(colors.RedText("Неверный выбор"))
		}
	}
}

func viewNotes(showFavorites bool) {
	var display []note.Note
	if showFavorites {
		display = storage.GetFavorites()
	} else {
		display = storage.GetNotes()
	}

	if len(display) == 0 {
		fmt.Println(colors.RedText("Список пуст"))
		return
	}

	const width = 86
	fmt.Println(colors.GreenText(strings.Repeat("-", width)))
	fmt.Printf(colors.GreenText("| %-3s | %-20s | %-30s | %-20s |\n"), "ID", "Заголовок", "Логин", "Пароль")
	fmt.Println(colors.GreenText(strings.Repeat("-", width)))

	for _, n := range display {
		title := n.DecryptTitle()
		if len(title) > 20 {
			title = title[:17] + "..."
		}
		fmt.Printf(colors.GreenText("| %-3d | %-20s | %-30s | %-20s |\n"),
			n.ID, title, n.Login, n.DecryptPassword())
	}
	fmt.Println(colors.GreenText(strings.Repeat("-", width)))
}

func EditCredentials(reader *bufio.Reader) {
	fmt.Println(colors.GreenText("\n--- Редактирование заметки ✏️ ---"))
	fmt.Println(colors.YellowText("Функция пока в разработке..."))
	// Можно реализовать позже
}

func DeleteNote(reader *bufio.Reader) {
	fmt.Println(colors.RedText("\n--- Удаление заметки ❌ ---"))
	fmt.Println(colors.YellowText("Функция пока в разработке..."))
	// Можно реализовать позже
}