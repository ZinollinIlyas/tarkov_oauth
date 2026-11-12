package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nickname    string     `gorm:"uniqueIndex;size:100"                           json:"nickname"`
	DisplayName string     `gorm:"size:100;not null"                              json:"display_name"`
	Password    string     `gorm:"column:password"                           json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	DisabledAt  *time.Time `json:"-"`
}

func (User) TableName() string {
	return "auth.users"
}

func (u *User) IsDisabled() bool {
	return u.DisabledAt != nil
}
