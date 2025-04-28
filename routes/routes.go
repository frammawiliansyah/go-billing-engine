package routes

import (
	"go-billing-engine/handlers"
	"go-billing-engine/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	pricingGroup := r.Group("/pricings")
	pricingGroup.Use(middlewares.AuthMiddleware())
	{
		pricingGroup.POST("/upsert", handlers.UpsertPricing)
	}

	loanGroup := r.Group("/loans")
	loanGroup.Use(middlewares.AuthMiddleware())
	{
		loanGroup.GET("/", handlers.GetAllLoans)
		loanGroup.POST("/", handlers.CreateLoan)
		loanGroup.GET("/:id", handlers.GetLoanDetail)
		loanGroup.GET("/oustanding/:id", handlers.GetOutstanding)
		loanGroup.POST("/payment/:loan_id", handlers.MakePayment)
		loanGroup.GET("/delinquent/:loan_id", handlers.IsDelinquent)
	}
}
