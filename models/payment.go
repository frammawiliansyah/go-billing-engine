package models

import "time"

type Payment struct {
	ID            uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint64    `gorm:"column:user_id;not null" json:"user_id"`
	LoanID        uint64    `gorm:"column:loan_id;not null" json:"loan_id"`
	PaymentCode   string    `gorm:"column:payment_code;type:varchar(255);uniqueIndex;not null" json:"payment_code"`
	PaymentAmount float64   `gorm:"column:payment_amount;type:numeric(20,2);not null" json:"payment_amount"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}
