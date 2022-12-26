package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ChatId    int64     `json:"chat_id"`
	MsgId     int       `json:"msg_id"`
	CreatedAt time.Time `json:"created_at"`
	Next      time.Time `json:"next"`
	Count     int       `json:"count"`
}

func NewTask(chatId int64, msgId int, delay time.Duration) *Task {
	return &Task{ChatId: chatId, MsgId: msgId, CreatedAt: time.Now(), Next: time.Now().Add(delay)}
}

// why string and not []byte? string is immutable

type Database interface {
	KeysWithPrefix(ctx context.Context, prefix string) ([]string, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type TaskDatabase struct {
	m        sync.Mutex
	lowLvlDb Database
}

func NewTaskDatabase(lowLvlDb Database) *TaskDatabase {
	return &TaskDatabase{lowLvlDb: lowLvlDb}
}

func (tdb *TaskDatabase) getKeyPrefix() string {
	return "/tasks/"
}
func (tdb *TaskDatabase) getTaskKey(task *Task) string {
	return tdb.getTaskKeyFromRaw(task.ChatId, task.MsgId)
}
func (tdb *TaskDatabase) getTaskKeyFromRaw(chatId int64, msgId int) string {
	return fmt.Sprintf("%s%d/%d", tdb.getKeyPrefix(), chatId, msgId)
}

func (tdb *TaskDatabase) GetAllTasks(ctx context.Context) ([]*Task, error) {
	keys, err := tdb.lowLvlDb.KeysWithPrefix(ctx, tdb.getKeyPrefix())
	if err != nil {
		return nil, fmt.Errorf("cant list keys: %w", err)
	}
	tasks := []*Task{}
	for _, key := range keys {
		task, err := tdb.getTaskByFullKey(ctx, key)
		if err != nil {
			return nil, err
		}

		if task == nil {
			return nil, fmt.Errorf("prefix return key '%s', but lowLvlDb does not contains it", key)
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}
func (tdb *TaskDatabase) Lock() {
	tdb.m.Lock()
}
func (tdb *TaskDatabase) UnLock() {
	tdb.m.Unlock()
}

func (tdb *TaskDatabase) GetTask(ctx context.Context, chatId int64, msgId int) (*Task, error) {
	return tdb.getTaskByFullKey(ctx, tdb.getTaskKeyFromRaw(chatId, msgId))
}

func (tdb *TaskDatabase) getTaskByFullKey(ctx context.Context, fullKey string) (*Task, error) {
	data, err := tdb.lowLvlDb.Get(ctx, fullKey)
	if err != nil {
		return nil, fmt.Errorf("cant get from lowLvlDb '%s': %w", fullKey, err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	task := &Task{}
	err = json.Unmarshal([]byte(data), task)
	if err != nil {
		return nil, fmt.Errorf("cant unmarshal from lowLvlDb '%s': %w", fullKey, err)
	}
	return task, nil
}
func (tdb *TaskDatabase) DeleteTask(ctx context.Context, task *Task) error {
	key := tdb.getTaskKey(task)
	err := tdb.lowLvlDb.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("cant delete from lowLvlDb '%s': %w", key, err)
	}
	return nil
}

func (tdb *TaskDatabase) SaveTask(ctx context.Context, task *Task) error {
	key := tdb.getTaskKey(task)
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("cant mashal task to lowLvlDb: %w", err)
	}
	err = tdb.lowLvlDb.Set(ctx, key, string(data))
	if err != nil {
		return fmt.Errorf("cant save to lowLvlDb '%s': %w", key, err)
	}
	return nil
}
