package tg

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Middleware interface {
	Income(ctx context.Context, msg *tgbotapi.Message) (next bool, err error)
}

type Ingress struct {
	middlewares []Middleware
}

func NewIngress(middlewares []Middleware) *Ingress {
	return &Ingress{middlewares: middlewares}
}

func (i *Ingress) Income(ctx context.Context, msg *tgbotapi.Message) error {
	for _, m := range i.middlewares {
		next, err := m.Income(ctx, msg)
		if err != nil {
			return err
		}
		if !next {
			return nil
		}
	}
	return nil
}
