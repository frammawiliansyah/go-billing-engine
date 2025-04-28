package handlers

import (
	"net/http"
	"time"

	"go-billing-engine/config"
	"go-billing-engine/models"

	"github.com/gin-gonic/gin"
)

func UpsertPricing(c *gin.Context) {
	var input struct {
		ID           uint64  `json:"id"`
		InterestRate float64 `json:"interest_rate" binding:"required"`
		AdminRate    float64 `json:"admin_rate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	var pricing models.Pricing
	if input.ID != 0 {
		if err := tx.First(&pricing, input.ID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Pricing not found"})
			return
		}

		pricing.InterestRate = input.InterestRate
		pricing.AdminRate = input.AdminRate
		pricing.UpdatedAt = time.Now()

		if err := tx.Save(&pricing).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pricing"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"message": "Pricing updated successfully", "pricing": pricing})
	} else {
		var existingPricing models.Pricing
		if err := tx.Where("interest_rate = ? AND admin_rate = ?", input.InterestRate, input.AdminRate).First(&existingPricing).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pricing with the same rates already exists"})
			return
		}

		newPricing := models.Pricing{
			InterestRate: input.InterestRate,
			AdminRate:    input.AdminRate,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := tx.Create(&newPricing).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusCreated, gin.H{"message": "Pricing created successfully", "pricing": newPricing})
	}
}
