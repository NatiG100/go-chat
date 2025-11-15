package usecases

import (
	"context"
	"encoding/json"
	"time"

	"example.com/go-chat/internal/core"
	"example.com/go-chat/internal/drivers"
	"github.com/google/uuid"
)

type ChatUsecase struct {
	repos core.Repositories
	rds   *drivers.RedisClient
}

func NewChatUsecase(r core.Repositories, rds *drivers.RedisClient) *ChatUsecase { return &ChatUsecase{repos: r, rds: rds} }

func (c *ChatUsecase) SendPrivate(ctx context.Context, from, to uuid.UUID, content string) (*core.Message, error) {
	m := &core.Message{SenderID: from, RecipientID: &to, Content: content, CreatedAt: time.Now()}
	if err := c.repos.MessageRepo().SaveMessage(ctx, m); err != nil {
		return nil, err
	}
	b, _ := json.Marshal(m)
	channel := "private:" + to.String()
	_ = c.rds.Publish(ctx, channel, string(b))
	return m, nil
}

func (c *ChatUsecase) SendGroup(ctx context.Context, from uuid.UUID, group uuid.UUID, content string) (*core.Message, error) {
	m := &core.Message{SenderID: from, GroupID: &group, Content: content, CreatedAt: time.Now()}
	if err := c.repos.MessageRepo().SaveMessage(ctx, m); err != nil {
		return nil, err
	}
	b, _ := json.Marshal(m)
	channel := "group:" + group.String()
	_ = c.rds.Publish(ctx, channel, string(b))
	return m, nil
}

func (c *ChatUsecase) GetPrivateHistory(ctx context.Context, a, b uuid.UUID, limit int) ([]core.Message, error) {
	return c.repos.MessageRepo().GetPrivateHistory(ctx, a, b, limit)
}

func (c *ChatUsecase) GetGroupHistory(ctx context.Context, group uuid.UUID, limit int) ([]core.Message, error) {
	return c.repos.MessageRepo().GetGroupHistory(ctx, group, limit)
}
