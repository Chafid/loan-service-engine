// main.go
package main

import (
	"log"
	"net/http"

	"loan-service-engine/config"
	"loan-service-engine/db"
	"loan-service-engine/handlers"
	"loan-service-engine/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Start the service")
	config.LoadEnv()
	db.Connect()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.POST("/login", handlers.Login)

	//Routes that needs authentications
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())
	api.GET("/loans/:id", handlers.GetLoanDetails)

	r.Static("/uploads", "./uploads")

	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		adminGroup.POST("/approve-loan", handlers.ApproveLoan)
		adminGroup.GET("/loan/:loan_id/agreement", handlers.DownloadLoanAgreement)
		adminGroup.POST("/disburse-loan", handlers.DisburseLoan)
		adminGroup.GET("/loans", handlers.ListLoans)

	}

	requesterGroup := api.Group("/requester")
	requesterGroup.Use(middleware.RequireRole("requester"))
	{
		requesterGroup.POST("/create-loan", handlers.CreateLoan)
	}

	investorGroup := api.Group("/investor")
	investorGroup.Use(middleware.RequireRole("investor"))
	{
		investorGroup.POST("/invest", handlers.InvestInLoan)
	}

	log.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}
