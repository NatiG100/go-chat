package core

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(100);not null" json:"username"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}


type Group struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// relationships
	Owner *User `gorm:"foreignKey:OwnerID;references:ID" json:"owner"`
}

type Message struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SenderID    uuid.UUID  `gorm:"type:uuid;not null" json:"sender_id"`
	RecipientID *uuid.UUID `gorm:"type:uuid" json:"recipient_id,omitempty"`
	GroupID     *uuid.UUID `gorm:"type:uuid" json:"group_id,omitempty"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// relationships
	Sender    *User  `gorm:"foreignKey:SenderID; references:ID" json:"sender"`
	Recipient *User  `gorm:"foreignKey:RecipientID; references:ID" json:"recipient"`
	Group     *Group `gorm:"foreignKey:GroupID; references:ID" json:"group"`
}


type GroupMember struct {
    GroupID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
    UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
    JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`

    // relationships
    Group *Group `gorm:"foreignKey:GroupID"`
    User  *User  `gorm:"foreignKey:UserID"`
}