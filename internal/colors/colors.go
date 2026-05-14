package colors

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func GreenText(text string) string {
	return colorGreen + text + colorReset
}

func RedText(text string) string {
	return colorRed + text + colorReset
}

func YellowText(text string) string {
	return colorYellow + text + colorReset
}