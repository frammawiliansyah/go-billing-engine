package models

import "time"

type Installment struct {
	ID                    uint64    `gorm:"primaryKey;column:id" json:"id"`
	LoanID                uint64    `gorm:"column:loan_id;not null" json:"loan_id"`
	UserID                uint64    `gorm:"column:user_id;not null" json:"user_id"`
	Sequence              int       `gorm:"column:sequence;not null" json:"sequence"`
	InstallmentAmount     float64   `gorm:"column:installment_amount;type:numeric(20,2);not null" json:"installment_amount"`
	InterestAmount        float64   `gorm:"column:interest_amount;type:numeric(20,2);not null" json:"interest_amount"`
	PrincipalAmount       float64   `gorm:"column:principal_amount;type:numeric(20,2);not null" json:"principal_amount"`
	OutstandingAmount     float64   `gorm:"column:outstanding_amount;type:numeric(20,2);not null" json:"outstanding_amount"`
	DueDate               time.Time `gorm:"column:due_date;not null" json:"due_date"`
	PaidStatus            string    `gorm:"column:paid_status;type:varchar(50);default:'PENDING';not null" json:"paid_status"`
	PaidAmountInstallment float64   `gorm:"column:paid_amount_installment;type:numeric(20,2);not null" json:"paid_amount_installment"`
	PaidAmountInterest    float64   `gorm:"column:paid_amount_interest;type:numeric(20,2);not null" json:"paid_amount_interest"`
	PaidAmountPrincipal   float64   `gorm:"column:paid_amount_principal;type:numeric(20,2);not null" json:"paid_amount_principal"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updated_at"`
}
