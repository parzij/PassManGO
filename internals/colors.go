package main

// Здесь убраны проверки темы (тёмная/светлая). Жёстко задаём цвета.

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// greenText возвращает строку с зелёным цветом
func greenText(text string) string {
	return colorGreen + text + colorReset
}

// redText возвращает строку с красным цветом
func redText(text string) string {
	return colorRed + text + colorReset
}

// yellowText возвращает строку с жёлтым цветом
func yellowText(text string) string {
	return colorYellow + text + colorReset
}
