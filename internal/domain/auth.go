package domain

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

type AuthMiddleware struct {
	db     Database
	secret string
}

func NewAuthMiddleware(db Database, secret string) *AuthMiddleware {
	return &AuthMiddleware{db: db, secret: secret}
}

type chatAuthInfo struct {
	Name  string `json:"name"`
	Allow bool   `json:"allow"`
}

func (a *AuthMiddleware) getKey(chat *tgbotapi.Chat) string {
	return fmt.Sprintf("/auth/chat/%d", chat.ID)
}
func (a *AuthMiddleware) findAuthInfo(ctx context.Context, chat *tgbotapi.Chat) (*chatAuthInfo, error) {
	data, err := a.db.Get(ctx, a.getKey(chat))
	if err != nil {
		return nil, fmt.Errorf("cant get auth from lowLvlDb: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	userInfo := &chatAuthInfo{}
	err = json.Unmarshal([]byte(data), userInfo)
	if err != nil {
		return nil, fmt.Errorf("cant unmarshall auth from lowLvlDb: %w", err)
	}
	return userInfo, nil

}
func (a *AuthMiddleware) saveAuthInfo(ctx context.Context, chat *tgbotapi.Chat, authInfo *chatAuthInfo) error {
	data, err := json.Marshal(authInfo)
	if err != nil {
		return fmt.Errorf("cant marshal auth to lowLvlDb: %w", err)
	}
	err = a.db.Set(ctx, a.getKey(chat), string(data))
	if err != nil {
		return fmt.Errorf("cant save auth to lowLvlDb: %w", err)
	}
	return nil
}

func (a *AuthMiddleware) Income(ctx context.Context, msg *tgbotapi.Message) (bool, error) {
	authInfo, err := a.findAuthInfo(ctx, msg.Chat)
	if err != nil {
		return false, err
	}
	if authInfo == nil {
		authInfo = &chatAuthInfo{}
		authInfo.Name = msg.Chat.UserName
	}

	if strings.TrimSpace(msg.Text) == a.secret {
		authInfo.Allow = true
		err = a.saveAuthInfo(ctx, msg.Chat, authInfo)
		if err != nil {
			return false, err
		}
		log.Printf("[INFO] verify ok '%s'", authInfo.Name)
		return false, nil
	}

	if !authInfo.Allow {
		log.Printf("[INFO] igonre '%s'", authInfo.Name)
	}
	return authInfo.Allow, nil
}
