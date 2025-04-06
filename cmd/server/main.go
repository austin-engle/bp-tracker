// File: cmd/server/main.go

package main

import (
	"context"
	"encoding/json"
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

	// --- API Endpoints for external clients (e.g., iOS app) ---
	apiGroup := router.Group("/api")
	{
		// Endpoint to get all readings as JSON
		apiGroup.GET("/readings", gin.WrapF(h.GetAllReadingsJSONHandler))
		// Add other future API endpoints here
		// Endpoint to get statistics as JSON
		apiGroup.GET("/stats", gin.WrapF(h.GetStatsHandler))
	}

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
	// Log the raw request event structure as JSON for detailed debugging
	requestBytes, _ := json.Marshal(req) // Ignoring marshalling error for logging purposes
	log.Printf("RAW REQUEST EVENT: %s\n", string(requestBytes))

	// Original logging (keeping it for now)
	log.Printf("Received request: Method=%s Path=%s\n", req.HTTPMethod, req.Path)

	// Call the Gin adapter
	resp, err := ginLambda.ProxyWithContext(ctx, req)

	// Log the response details
	responseBytes, _ := json.Marshal(resp) // Ignoring marshalling error for logging purposes
	log.Printf("RAW RESPONSE: %s\n", string(responseBytes))
	log.Printf("Sending response: StatusCode=%d Body=%.50s... Error=%v\n", resp.StatusCode, resp.Body, err)

	return resp, err
}

func main() {
	// Start the Lambda listener
	lambda.Start(LambdaHandler)
}
