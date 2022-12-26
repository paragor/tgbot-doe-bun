package domain

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

type DelayIncomeMiddleware struct {
	db    *TaskDatabase
	tg    *tgbotapi.BotAPI
	delay time.Duration
}

func NewDelayIncomeMiddleware(db *TaskDatabase, tg *tgbotapi.BotAPI, delay time.Duration) *DelayIncomeMiddleware {
	return &DelayIncomeMiddleware{db: db, tg: tg, delay: delay}
}

func (p *DelayIncomeMiddleware) Income(ctx context.Context, msg *tgbotapi.Message) (bool, error) {
	p.db.Lock()
	defer p.db.UnLock()

	task := NewTask(msg.Chat.ID, msg.MessageID, p.delay)
	saveTaskErr := p.db.SaveTask(ctx, task)
	if saveTaskErr != nil {
		answer := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("ERROR %s", saveTaskErr.Error()))
		answer.ReplyToMessageID = msg.MessageID
		if _, err := p.tg.Send(answer); err != nil {
			return false, fmt.Errorf("done: cant send answer: %s. Origin err: %w", err.Error(), saveTaskErr)
		}
		return false, saveTaskErr

	}
	answer := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("ok, notify after %s. Reply origin msg with txt 'done' to finish it.", task.Next.Format(time.RFC822)))
	answer.ReplyToMessageID = msg.MessageID
	if _, err := p.tg.Send(answer); err != nil {
		return false, fmt.Errorf("done: cant send 'ok' answer: %s", err)
	}

	return true, nil
}
