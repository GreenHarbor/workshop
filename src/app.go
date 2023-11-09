package main

import (
	"fmt"
	"log"
	"net/http"

	// "workshop/helpers"
	"workshop/routes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
)

var svc *dynamodb.DynamoDB
var tableName = "workshop"

func main() {
	// var getKeyResult helpers.AWSCredentials
	// getKeyResult, err := helpers.GetAccessKeys()
	// if err != nil {
	// 	log.Println("Error getting access keys")
	// 	log.Fatal(err)
	// }

	AKID := ""
	SECRET_KEY := ""
	// AKID := getKeyResult.AWSAccessKeyID
	// SECRET_KEY := getKeyResult.AWSSecretAccessKey

	log.Println("Here is the AKID")
	log.Println(AKID)
	log.Println("Here is the SECRET_KEY")
	log.Println(SECRET_KEY)

	// Initialize a session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewStaticCredentials(AKID, SECRET_KEY, ""),
	})
	if err != nil {
		log.Println("Error getting session:")
		log.Fatal(err)
	}
	fmt.Println("After creating session")

	// Create DynamoDB client and expose HTTP requests/responses
	svc = dynamodb.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))

	//routes
	r := mux.NewRouter()
	routes.RegisterRoutes(r, svc, tableName)

	if http.ListenAndServe(":8080", r) != nil {
		log.Fatalf("Failed to create server at port 8080")
	}
}
