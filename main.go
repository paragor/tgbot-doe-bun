package main

import (
	"context"
	"github.com/paragor/tgbot-doe-bun/internal/badgerdb"
	"github.com/paragor/tgbot-doe-bun/internal/domain"
	"github.com/paragor/tgbot-doe-bun/internal/tg"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	badgerDb := os.Getenv("BADGER_DB")
	if badgerDb == "" {
		panic("set env 'BADGER_DB' path for badger db")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		panic("set env 'SECRET_KEY'")
	}

	telegramapitoken := os.Getenv("TELEGRAM_APITOKEN")
	if telegramapitoken == "" {
		panic("set env 'TELEGRAM_APITOKEN'")
	}
	bot, err := tgbotapi.NewBotAPI(telegramapitoken)
	if err != nil {
		panic(err)
	}

	lowLevelDb, closeDb, err := badgerdb.NewBadgerDb(badgerDb)
	if err != nil {
		panic("cant create badger: " + err.Error())
	}
	defer closeDb()
	tdb := domain.NewTaskDatabase(lowLevelDb)
	delayTime := time.Hour
	ingress := tg.NewIngress([]tg.Middleware{
		domain.NewAuthMiddleware(lowLevelDb, secretKey),
		domain.NewDoneMiddleware(tdb, bot),
		domain.NewDelayIncomeMiddleware(tdb, bot, delayTime),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notify := domain.NewNotifyWorker(time.Minute, delayTime, tdb, bot)
	go func() {
		err := notify.Run()
		if err != nil {
			log.Printf("[ERROR] notify: %s\n", err.Error())
			cancel()
		}
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)
	log.Println("listen updates...")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		err := ingress.Income(ctx, update.Message)
		if err != nil {
			log.Printf("[ERROR] msg: '%s'. err: %s. who: %s\n", update.Message.Text, err, update.Message.Chat.UserName)
		} else {
			log.Printf("[DEBUG] OK msg: '%s'. who: %s\n", update.Message.Text, update.Message.Chat.UserName)
		}
		select {
		case <-ctx.Done():
			log.Fatal("done signal")
		default:
		}
	}
}
