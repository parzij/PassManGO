package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

var (
	shutdownTimer   *time.Timer
	warningTimer    *time.Timer
	shutdownEnabled = true
)

func initShutdownTimers() {
	// Устанавливаем таймер на 2 минуты
	shutdownTimer = time.AfterFunc(2*time.Minute, func() {
		fmt.Printf("\n%sВремя сеанса истекло. Программа будет закрыта.%s\n", colorRed, colorReset)
		os.Exit(0)
	})

	// Устанавливаем предупреждение за 1 минуту до закрытия
	warningTimer = time.AfterFunc(1*time.Minute, func() {
		fmt.Printf("\n%sВнимание! Программа будет автоматически закрыта через 1 минуту.%s\n", colorYellow, colorReset)
	})

	// Обработка Ctrl+C для отключения авто-закрытия
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		shutdownTimer.Stop()
		warningTimer.Stop()
		shutdownEnabled = false
		fmt.Printf("\n%sАвтоматическое закрытие отключено. Программа не будет закрыта автоматически.%s\n", colorGreen, colorReset)
	}()
}

func resetShutdownTimer() {
	if shutdownEnabled {
		shutdownTimer.Reset(2 * time.Minute)
		warningTimer.Reset(1 * time.Minute)
	}
}

func greenText(text string) string {
	return colorGreen + text + colorReset
}

func redText(text string) string {
	return colorRed + text + colorReset
}

func yellowText(text string) string {
	return colorYellow + text + colorReset
}
