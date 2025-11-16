package core

import (
	"context"

	"github.com/google/uuid"
)

// small repository interfaces used by usecases

type UserRepository interface {
	CreateUser(ctx context.Context, u *User, plainPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	VerifyPassword(ctx context.Context, email, plain string) (*User, error)
}

type GroupRepository interface {
	CreateGroup(ctx context.Context, g *Group) error
	MyGroups(ctx context.Context, userId uuid.UUID) ([]Group, error)
	AddGroupMember(ctx context.Context, groupID, userID uuid.UUID) error
	ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]User, error)
}

type MessageRepository interface {
	SaveMessage(ctx context.Context, m *Message) error
	GetPrivateHistory(ctx context.Context, a, b uuid.UUID, limit int) ([]Message, error)
	GetGroupHistory(ctx context.Context, groupID uuid.UUID, limit int) ([]Message, error)
}

// Repositories groups
type Repositories interface {
	UserRepo() UserRepository
	GroupRepo() GroupRepository
	MessageRepo() MessageRepository
}
