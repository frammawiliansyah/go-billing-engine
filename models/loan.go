package models

import "time"

type Loan struct {
	ID         uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint64    `gorm:"column:user_id;not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	PricingID  uint64    `gorm:"column:pricing_id;not null" json:"pricing_id"`
	Pricing    Pricing   `gorm:"foreignKey:PricingID" json:"pricing"`
	LoanCode   string    `gorm:"column:loan_code;type:varchar(255);uniqueIndex;not null" json:"loan_code"`
	LoanStatus string    `gorm:"column:loan_status;type:varchar(255);default:PENDING;not null" json:"loan_status"`
	LoanAmount float64   `gorm:"column:loan_amount;type:numeric(20,2);not null" json:"loan_amount"`
	LoanLength int       `gorm:"column:loan_length;type:integer;not null" json:"loan_length"`
	NTFTotal   float64   `gorm:"column:ntf_total;type:numeric(20,2)" json:"ntf_total"`
	AdminTotal float64   `gorm:"column:admin_total;type:numeric(20,2)" json:"admin_total"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}
