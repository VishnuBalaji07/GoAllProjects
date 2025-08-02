package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Password  string         `json:"-"` // omit password in JSON output
	Phone     string         `json:"phone"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Chat represents a conversation (group or one-on-one)
type Chat struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `json:"name,omitempty"`
	Description   string         `json:"description,omitempty"`
	IsGroup       bool           `json:"is_group"`
	CreatedBy     uint           `json:"created_by"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	LastMessage   *string        `json:"last_message,omitempty" `
	LastUpdatedAt *time.Time     `json:"last_updated_at,omitempty"`
	Members       []ChatMember   `json:"members" gorm:"foreignKey:ChatID"`
	Messages      []Message      `json:"messages,omitempty" gorm:"foreignKey:ChatID"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatMember links users to chats with additional metadata
type ChatMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	ChatID   uint      `json:"chat_id"`
	UserID   uint      `json:"user_id"`
	AddedBy  *uint     `json:"added_by,omitempty"`
	Role     string    `json:"role,omitempty"` // e.g. "admin", "member"
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
	// Chat field removed for clarity unless specifically required
}

// Message represents a message sent in a chat
type Message struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	ChatID      uint            `gorm:"index" json:"chat_id"`
	SenderID    uint            `gorm:"index" json:"sender_id"`
	Text        string          `json:"text"`
	Type        string          `json:"type"` // e.g., "text", "image"
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	ReplyToID   *uint           `gorm:"index" json:"reply_to_id,omitempty"`
	ReplyTo     *Message        `gorm:"foreignKey:ReplyToID" json:"-"`
	Reactions   []Reaction      `gorm:"foreignKey:MessageID" json:"reactions"`
	StatusTrack []MessageStatus `gorm:"foreignKey:MessageID" json:"status_track"`
	Sender      *User           `json:"sender,omitempty"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// MessageStatus tracks whether a message has been delivered/read per user
type MessageStatus struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	MessageID    uint       `gorm:"index" json:"message_id"`
	UserID       uint       `gorm:"index" json:"-"`
	Status       string     `json:"status"` // e.g. "sent", "delivered", "read"
	SentAt       *time.Time `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	ReadAt       *time.Time `json:"read_at,omitempty"`
	ChatMemberID uint       `gorm:"index" json:"chat_member_id"`
}

// Reaction stores emoji reactions on messages
type Reaction struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	MessageID uint   `gorm:"index" json:"message_id"`
	Emoji     string `json:"emoji"`
	UserID    uint   `gorm:"index" json:"-"`
}
