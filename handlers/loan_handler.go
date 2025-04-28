package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-billing-engine/config"
	"go-billing-engine/models"
	"go-billing-engine/utils"

	"github.com/gin-gonic/gin"
)

type LoanListDTO struct {
	ID         uint64  `json:"id"`
	UserID     uint64  `json:"user_id"`
	LoanCode   string  `json:"loan_code"`
	LoanStatus string  `json:"loan_status"`
	LoanAmount float64 `json:"loan_amount"`
	LoanLength int     `json:"loan_length"`
	NTFTotal   float64 `json:"ntf_total"`
	CreatedAt  string  `json:"created_at"`
}

type InstallmentDTO struct {
	Sequence          int       `json:"sequence"`
	InstallmentAmount float64   `json:"installment_amount"`
	InterestAmount    float64   `json:"interest_amount"`
	PrincipalAmount   float64   `json:"principal_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	DueDate           time.Time `json:"due_date"`
	PaidStatus        string    `json:"paid_status"`
}

type LoanDetailDTO struct {
	ID           uint64           `json:"id"`
	UserID       uint64           `json:"user_id"`
	LoanCode     string           `json:"loan_code"`
	LoanStatus   string           `json:"loan_status"`
	LoanAmount   float64          `json:"loan_amount"`
	LoanLength   int              `json:"loan_length"`
	NTFTotal     float64          `json:"ntf_total"`
	AdminTotal   float64          `json:"admin_total"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Installments []InstallmentDTO `json:"installments"`
}

func GetAllLoans(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	var loans []models.Loan
	if err := config.DB.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&loans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch loans"})
		return
	}

	var loanList []LoanListDTO
	for _, loan := range loans {
		loanList = append(loanList, LoanListDTO{
			ID:         loan.ID,
			UserID:     loan.UserID,
			LoanCode:   loan.LoanCode,
			LoanStatus: loan.LoanStatus,
			LoanAmount: loan.LoanAmount,
			LoanLength: loan.LoanLength,
			NTFTotal:   loan.NTFTotal,
			CreatedAt:  loan.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Loans fetched successfully",
		"page":    page,
		"limit":   limit,
		"loans":   loanList,
	})
}

func GetLoanDetail(c *gin.Context) {
	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseUint(loanIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	var loan models.Loan
	if err := config.DB.
		First(&loan, loanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	var installments []models.Installment
	if err := config.DB.
		Where("loan_id = ?", loan.ID).
		Order("sequence asc").
		Find(&installments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load installments"})
		return
	}

	var installmentDTOs []InstallmentDTO
	for _, inst := range installments {
		installmentDTOs = append(installmentDTOs, InstallmentDTO{
			Sequence:          inst.Sequence,
			InstallmentAmount: inst.InstallmentAmount,
			InterestAmount:    inst.InterestAmount,
			PrincipalAmount:   inst.PrincipalAmount,
			OutstandingAmount: inst.OutstandingAmount,
			DueDate:           inst.DueDate,
			PaidStatus:        inst.PaidStatus,
		})
	}

	response := LoanDetailDTO{
		ID:           loan.ID,
		UserID:       loan.UserID,
		LoanCode:     loan.LoanCode,
		LoanStatus:   loan.LoanStatus,
		LoanAmount:   loan.LoanAmount,
		LoanLength:   loan.LoanLength,
		NTFTotal:     loan.NTFTotal,
		AdminTotal:   loan.AdminTotal,
		CreatedAt:    loan.CreatedAt,
		UpdatedAt:    loan.UpdatedAt,
		Installments: installmentDTOs,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Loan detail fetched successfully",
		"loan":    response,
	})
}

func CreateLoan(c *gin.Context) {
	var input struct {
		LoanAmount float64 `json:"loan_amount" binding:"required"`
		LoanLength int     `json:"loan_length" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDFloat, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := uint64(userIDFloat.(float64))

	var existingLoan models.Loan
	if err := config.DB.Where("user_id = ? AND loan_status != ?", userID, "CLOSED").First(&existingLoan).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has an active loan, cannot create new loan"})
		return
	}

	var pricing models.Pricing
	if err := config.DB.First(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No pricing found"})
		return
	}

	loanAmount := input.LoanAmount
	loanLength := input.LoanLength
	adminTotal := loanAmount * (pricing.AdminRate / 100)
	ntfTotal := loanAmount + adminTotal
	annualInterestRate := pricing.InterestRate / 100
	weeklyInterestRate := annualInterestRate / 52
	pmt := utils.PMT(weeklyInterestRate, loanLength, ntfTotal)

	tx := config.DB.Begin()

	generatedLoanCode := utils.GenerateLoanCode()

	loan := models.Loan{
		UserID:     userID,
		PricingID:  pricing.ID,
		LoanCode:   generatedLoanCode,
		LoanStatus: "ACTIVE",
		LoanAmount: loanAmount,
		LoanLength: loanLength,
		NTFTotal:   ntfTotal,
		AdminTotal: adminTotal,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := tx.Create(&loan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create loan"})
		return
	}

	baseDate := time.Now().AddDate(0, 0, 7)

	currentNTFTotal := ntfTotal

	for i := 1; i <= loanLength; i++ {
		interest := currentNTFTotal * weeklyInterestRate
		principal := pmt - interest
		currentNTFTotal -= principal

		installment := models.Installment{
			LoanID:                loan.ID,
			UserID:                loan.UserID,
			Sequence:              i,
			InstallmentAmount:     pmt,
			InterestAmount:        interest,
			PrincipalAmount:       principal,
			OutstandingAmount:     utils.Max(currentNTFTotal, 0),
			DueDate:               baseDate.AddDate(0, 0, (i-1)*7),
			PaidStatus:            "PENDING",
			PaidAmountInstallment: 0,
			PaidAmountInterest:    0,
			PaidAmountPrincipal:   0,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		if err := tx.Create(&installment).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create installment"})
			return
		}
	}

	tx.Commit()

	if err := config.DB.Preload("User").Preload("Pricing").First(&loan, loan.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load loan details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Loan created successfully",
		"loan":    loan,
	})
}

func GetOutstanding(c *gin.Context) {
	loanID := c.Param("id")

	var loan models.Loan
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	var installments []models.Installment
	if err := config.DB.
		Where("loan_id = ? AND paid_status = ?", loanID, "PENDING").
		Order("sequence asc").
		Find(&installments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load installments"})
		return
	}

	var principalTotal float64
	var interestTotal float64
	var overdueCount int

	today := time.Now()

	for _, inst := range installments {
		remainingPrincipal := inst.PrincipalAmount - inst.PaidAmountPrincipal
		remainingInterest := inst.InterestAmount - inst.PaidAmountInterest

		principalTotal += remainingPrincipal
		interestTotal += remainingInterest

		if inst.DueDate.Before(today) {
			overdueCount++
		}
	}

	if principalTotal > loan.NTFTotal {
		diff := principalTotal - loan.NTFTotal
		interestTotal += diff
		principalTotal -= diff
	}

	var dueDate interface{} = nil
	if len(installments) > 0 {
		dueDate = installments[0].DueDate
	}

	outstandingInformation := gin.H{
		"outstanding_installment": utils.RoundFloat(principalTotal+interestTotal, 2),
		"outstanding_interest":    utils.RoundFloat(interestTotal, 2),
		"outstanding_principal":   utils.RoundFloat(principalTotal, 2),
		"due_date":                dueDate,
		"is_delinquent":           overdueCount >= 2,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                 "Outstanding fetched successfully",
		"outstanding_information": outstandingInformation,
	})
}

func MakePayment(c *gin.Context) {
	loanID := c.Param("loan_id")

	var input struct {
		PaymentAmount float64 `json:"payment_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var loan models.Loan
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	var installments []models.Installment
	if err := config.DB.
		Where("loan_id = ? AND paid_status = ?", loanID, "PENDING").
		Order("sequence asc").
		Find(&installments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load installments"})
		return
	}

	if len(installments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending installments"})
		return
	}

	var totalOutstanding float64
	for _, inst := range installments {
		remainingPrincipal := inst.PrincipalAmount - inst.PaidAmountPrincipal
		remainingInterest := inst.InterestAmount - inst.PaidAmountInterest
		totalOutstanding += remainingPrincipal + remainingInterest
	}

	if input.PaymentAmount > totalOutstanding {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Payment amount %.2f exceeds total outstanding %.2f", input.PaymentAmount, totalOutstanding),
		})
		return
	}

	var overdueInstallmentTotal float64
	today := time.Now()

	for _, inst := range installments {
		daysOverdue := today.Sub(inst.DueDate).Hours() / 24
		if daysOverdue >= 14 {
			overdueInstallmentTotal += inst.InstallmentAmount
		}
	}

	if overdueInstallmentTotal > 0 && input.PaymentAmount < overdueInstallmentTotal {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Payment must cover at least overdue amount: %.2f", overdueInstallmentTotal),
		})
		return
	}

	remainingAmount := input.PaymentAmount

	for i := 0; i < len(installments) && remainingAmount > 0; i++ {
		inst := &installments[i]

		payPrincipal := inst.PrincipalAmount - inst.PaidAmountPrincipal
		payInterest := inst.InterestAmount - inst.PaidAmountInterest
		payInstallment := payPrincipal + payInterest

		if remainingAmount >= payInstallment {
			inst.PaidAmountPrincipal = inst.PrincipalAmount
			inst.PaidAmountInterest = inst.InterestAmount
			inst.PaidAmountInstallment = inst.InstallmentAmount
			inst.PaidStatus = "PAID"
			remainingAmount -= payInstallment
		} else {
			ratio := remainingAmount / payInstallment
			inst.PaidAmountPrincipal += payPrincipal * ratio
			inst.PaidAmountInterest += payInterest * ratio
			inst.PaidAmountInstallment += payInstallment * ratio
			remainingAmount = 0
		}

		inst.UpdatedAt = time.Now()

		if err := config.DB.Save(inst).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update installment"})
			return
		}
	}

	payment := models.Payment{
		UserID:        loan.UserID,
		LoanID:        loan.ID,
		PaymentCode:   utils.GeneratePaymentCode(),
		PaymentAmount: input.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	var pendingCount int64
	if err := config.DB.Model(&models.Installment{}).
		Where("loan_id = ? AND paid_status = ?", loanID, "PENDING").
		Count(&pendingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check loan status"})
		return
	}

	if pendingCount == 0 {
		loan.LoanStatus = "CLOSED"
		loan.UpdatedAt = time.Now()

		if err := config.DB.Save(&loan).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close loan"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment processed successfully",
		"payment": payment,
	})
}

func IsDelinquent(c *gin.Context) {
	loanID := c.Param("loan_id")

	var loan models.Loan
	if err := config.DB.First(&loan, loanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	var pendingInstallments []models.Installment
	if err := config.DB.
		Where("loan_id = ? AND paid_status = ?", loanID, "PENDING").
		Order("sequence asc").
		Find(&pendingInstallments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installments"})
		return
	}

	today := time.Now()
	var overdueInstallments []models.Installment
	var minimumPayment float64

	for _, inst := range pendingInstallments {
		if today.Sub(inst.DueDate).Hours()/24 >= 14 {
			overdueInstallments = append(overdueInstallments, inst)
			minimumPayment += inst.InstallmentAmount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Delinquent installments fetched successfully",
		"data": gin.H{
			"minimum_payment_required": utils.RoundFloat(minimumPayment, 2),
			"overdue_installments":     overdueInstallments,
		},
	})
}
