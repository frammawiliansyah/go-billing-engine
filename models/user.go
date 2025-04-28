package models

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	FullName     string    `gorm:"column:full_name;type:varchar(255)" json:"full_name"`
	EmailAddress string    `gorm:"column:email_address;type:varchar(255);uniqueIndex" json:"email_address"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	PasswordSalt string    `gorm:"column:password_salt;type:varchar(255)" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}
