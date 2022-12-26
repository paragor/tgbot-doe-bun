package domain

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

type DoneMiddleware struct {
	db *TaskDatabase
	tg *tgbotapi.BotAPI
}

func NewDoneMiddleware(db *TaskDatabase, tg *tgbotapi.BotAPI) *DoneMiddleware {
	return &DoneMiddleware{db: db, tg: tg}
}

func (p *DoneMiddleware) Income(ctx context.Context, msg *tgbotapi.Message) (next bool, getTaskErr error) {
	if strings.TrimSpace(msg.Text) != "done" {
		return true, nil
	}
	if msg.ReplyToMessage == nil {
		answer := tgbotapi.NewMessage(msg.Chat.ID, "ERROR you should reply what is done")
		answer.ReplyToMessageID = msg.MessageID
		if _, err := p.tg.Send(answer); err != nil {
			return false, fmt.Errorf("done: cant send answer: %w", err)
		}
		return false, nil
	}
	p.db.Lock()
	defer p.db.UnLock()

	task, getTaskErr := p.db.GetTask(ctx, msg.Chat.ID, msg.ReplyToMessage.MessageID)
	if getTaskErr != nil {
		answer := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("ERROR %s", getTaskErr.Error()))
		answer.ReplyToMessageID = msg.MessageID
		if _, err := p.tg.Send(answer); err != nil {
			return false, fmt.Errorf("done: cant send answer: %s. Origin err: %w", err.Error(), getTaskErr)
		}
		return false, getTaskErr
	}

	if task == nil {
		originErr := fmt.Errorf("cant found task")
		answer := tgbotapi.NewMessage(msg.Chat.ID, "WARN not found task")
		answer.ReplyToMessageID = msg.MessageID
		if _, err := p.tg.Send(answer); err != nil {
			return false, fmt.Errorf("done: cant send answer: %s. Origin err: %w", err.Error(), originErr)
		}
		return false, originErr
	}
	originErr := p.db.DeleteTask(ctx, task)
	if originErr != nil {
		answer := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("ERRR cant delete task: %s", originErr.Error()))
		answer.ReplyToMessageID = msg.MessageID
		if _, err := p.tg.Send(answer); err != nil {
			return false, fmt.Errorf("delete: cant send answer: %s. Origin err: %w", err.Error(), originErr)
		}
		return false, originErr
	}
	answer := tgbotapi.NewMessage(msg.Chat.ID, "OK, done")
	answer.ReplyToMessageID = msg.MessageID
	if _, err := p.tg.Send(answer); err != nil {
		return false, fmt.Errorf("done: cant send ok answer:%w", err)
	}

	return false, nil
}
