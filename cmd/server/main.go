// File: cmd/server/main.go

package main

import (
	"context"
	"log"
	"net/http" // Needed for http.Dir and handler funcs

	"bp-tracker/internal/database"
	"bp-tracker/internal/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

// Middleware wrapper for Gin
func ginSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

func init() {
	log.Println("Initializing Lambda handler...")

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize database: %v", err)
	}

	// Initialize handlers (needed for handler methods)
	h, err := handlers.New(db)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize handlers: %v", err)
	}

	// Create a Gin engine - Define routes directly in Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginSecurityHeaders()) // Apply security headers middleware globally

	// Define routes using gin.WrapF for the http.HandlerFunc methods
	router.GET("/", gin.WrapF(h.HomeHandler))
	router.POST("/submit", gin.WrapF(h.SubmitReadingHandler))
	router.GET("/export/csv", gin.WrapF(h.ExportCSVHandler))

	// Use POST for potentially state-changing operation
	router.POST("/migrate", gin.WrapF(h.MigrateHandler))

	// Serve static files using Gin's StaticFS
	// Note: Ensure the path in the Dockerfile copies web/static correctly
	// The path served ("/static") must match links in HTML
	// The second path is the filesystem path *inside the container*
	router.StaticFS("/static", http.Dir("web/static"))

	// Create the Gin Lambda adapter with the Gin router
	ginLambda = ginadapter.New(router)
	log.Println("Lambda handler initialized successfully (using direct Gin routing)")
}

// LambdaHandler is the main handler function for AWS Lambda
func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: Method=%s Path=%s\n", req.HTTPMethod, req.Path)
	resp, err := ginLambda.ProxyWithContext(ctx, req)
	log.Printf("Sending response: StatusCode=%d Body=%.50s... Error=%v\n", resp.StatusCode, resp.Body, err)
	return resp, err
}

func main() {
	// Start the Lambda listener
	lambda.Start(LambdaHandler)
}
