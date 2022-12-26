package domain

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type NotifyWorker struct {
	every time.Duration
	tdb   *TaskDatabase
	tg    *tgbotapi.BotAPI
}

func NewNotifyWorker(every time.Duration, tdb *TaskDatabase, tg *tgbotapi.BotAPI) *NotifyWorker {
	return &NotifyWorker{every: every, tdb: tdb, tg: tg}
}

func (n *NotifyWorker) Run() error {
	for {
		err := n.round()
		if err != nil {
			return err
		}
		time.Sleep(n.every)
	}
}

func (n *NotifyWorker) round() error {
	ctx := context.TODO()
	n.tdb.Lock()
	defer n.tdb.UnLock()
	tasks, err := n.tdb.GetAllTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if !time.Now().After(task.Next) {
			continue
		}
		log.Printf("[DEBUG] notify: %#v", *task)
		msg := tgbotapi.NewMessage(task.ChatId, fmt.Sprintf("^ %d", task.Count))
		msg.ReplyToMessageID = task.MsgId
		if _, err := n.tg.Send(msg); err != nil {
			return fmt.Errorf("notify: cant send msg: %w", err)
		}
		task.Count++
		err = n.tdb.SaveTask(ctx, task)
		if err != nil {
			return fmt.Errorf("save: cant send task: %w", err)
		}
	}
	return nil
}
