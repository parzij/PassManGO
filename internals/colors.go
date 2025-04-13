package main

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
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
