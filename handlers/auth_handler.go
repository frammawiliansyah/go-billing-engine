package handlers

import (
	"net/http"
	"strings"

	"go-billing-engine/config"
	"go-billing-engine/models"
	"go-billing-engine/utils"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input struct {
		FullName     string `json:"full_name" binding:"required"`
		EmailAddress string `json:"email_address" binding:"required,email"`
		Password     string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := config.DB.Where("email_address = ?", strings.ToLower(input.EmailAddress)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already registered"})
		return
	}

	salt, err := utils.GenerateRandomSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate salt"})
		return
	}
	passwordHash := utils.HashPassword(input.Password, salt)

	user := models.User{
		FullName:     strings.ToUpper(input.FullName),
		EmailAddress: strings.ToLower(input.EmailAddress),
		PasswordHash: passwordHash,
		PasswordSalt: salt,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var input struct {
		EmailAddress string `json:"email_address" binding:"required,email"`
		Password     string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email_address = ?", strings.ToLower(input.EmailAddress)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPassword(input.Password, user.PasswordSalt, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":            user.ID,
			"full_name":     user.FullName,
			"email_address": user.EmailAddress,
		},
		"token": token,
	})
}
