package models

import "time"

type Pricing struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`
	InterestRate float64   `gorm:"column:interest_rate;type:numeric(20,2)" json:"interest_rate"`
	AdminRate    float64   `gorm:"column:admin_rate;type:numeric(20,2)" json:"admin_rate"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}
