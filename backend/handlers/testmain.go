package handlers

import (
	"os"
	"testing"

	"github.com/junhwi/gobco"
)

func TestMain(m *testing.M) {
	// запускаємо звичайні тести
	code := m.Run()
	// збираємо branch coverage
	gobco.ReportCoverage()       // виведе статистику в консоль
	gobco.ReportProfile("c.out") // запише профіль у c.out
	os.Exit(code)
}
