package telegram

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

// Start запускає бота і планувальник
func Start(db *sqlx.DB) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("TELEGRAM_BOT_TOKEN not set, skipping Telegram bot")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create botAPI: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Запускаємо планувальник
	s := gocron.NewScheduler(time.Local)

	// Щоденне нагадування о 20:00
	s.Every(1).Day().At("20:00").Do(func() {
		sendDailyReminder(bot, db)
	})
	// ТЕСТ нагадування кожні 10сек
	s.Every(10).Second().Do(func() { sendDailyReminder(bot, db) })

	// Щотижневий звіт кожного понеділка о 09:00
	s.Every(1).Week().Monday().At("09:00").Do(func() {
		sendWeeklyReport(bot, db)
	})
	// ТЕСТ звіт кожні 30сек
	s.Every(30).Second().Do(func() {
		sendWeeklyReport(bot, db)
	})

	s.StartAsync()
}

// sendDailyReminder знаходить користувачів, які не додали сьогоднішній настрій, і надсилає їм повідомлення
func sendDailyReminder(bot *tgbotapi.BotAPI, db *sqlx.DB) {
	const query = `
        SELECT telegram_chat_id
        FROM users
        WHERE telegram_chat_id IS NOT NULL
          AND id NOT IN (
            SELECT user_id FROM mood WHERE date = CURRENT_DATE
          )`
	var chatIDs []int64
	if err := db.Select(&chatIDs, query); err != nil {
		log.Printf("sendDailyReminder db error: %v", err)
		return
	}

	for _, chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, "Не забудь внести сьогоднішній настрій")
		if _, err := bot.Send(msg); err != nil {
			log.Printf("failed to send daily reminder to %d: %v", chatID, err)
		}
	}
}

// sendWeeklyReport збирає статистику за попередній тиждень і надсилає її користувачам із зареєстрованим чат-ID
func sendWeeklyReport(bot *tgbotapi.BotAPI, db *sqlx.DB) {
	const usersQuery = `SELECT telegram_chat_id, id FROM users WHERE telegram_chat_id IS NOT NULL`
	type userRec struct {
		ChatID int64  `db:"telegram_chat_id"`
		UserID string `db:"id"`
	}
	var users []userRec
	if err := db.Select(&users, usersQuery); err != nil {
		log.Printf("sendWeeklyReport users err: %v", err)
		return
	}

	for _, u := range users {
		const statsQuery = `
            SELECT icon, COUNT(*) AS cnt
            FROM mood
            WHERE user_id = $1
              AND date >= CURRENT_DATE - INTERVAL '7 days'
            GROUP BY icon`
		rows, err := db.Queryx(statsQuery, u.UserID)
		if err != nil {
			log.Printf("stats query err for %s: %v", u.UserID, err)
			continue
		}
		text := "Твій звіт за останній тиждень:\n"
		for rows.Next() {
			var icon string
			var cnt int
			rows.Scan(&icon, &cnt)
			text += fmt.Sprintf("%s — %d\n", icon, cnt)
		}
		msg := tgbotapi.NewMessage(u.ChatID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("failed to send weekly report to %d: %v", u.ChatID, err)
		}
	}
}
