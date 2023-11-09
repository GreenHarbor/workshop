package tests

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"workshop/models"
	"workshop/routes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
)

var testRouter *mux.Router
var testServer *httptest.Server
var svc *dynamodb.DynamoDB
var tableName = "workshop_test"
var testDBSeedData = []models.Workshop{
	{
		Creator_Id:            "2",
		Creation_Timestamp:    "2023-10-20-21:22:22.080",
		Title:                 "Herbs Galore!",
		Description:           "Repurpose kitchen scraps to grow your own herbs with Ms Rafidah!",
		Location:              "123 Circle Road",
		Vacancies:             11,
		Attendees:             []string{"64"},
		Registration_Deadline: "2024-02-17-23:59:59.000",
		Start_Timestamp:       "2024-02-10-15:00:00.000",
	},
	{
		Creator_Id:            "1",
		Creation_Timestamp:    "2023-11-04-03:28:10.244",
		Title:                 "Mr Lee fan repair!",
		Description:           "Repair spoiled fans with Mr Lee",
		Location:              "123 Example Road",
		Vacancies:             7,
		Attendees:             []string{"999"},
		Registration_Deadline: "2024-02-08-23:59:59.000",
		Start_Timestamp:       "2024-02-15-15:00:00.000",
	},
}

func doesTableExist(name string, svc *dynamodb.DynamoDB) bool {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}
	_, err := svc.DescribeTable(input)
	if err != nil {
		return false
	} else {
		return true
	}
}

func hasTableBeenFullyDeleted(name string, svc *dynamodb.DynamoDB) bool {
	maxRetries := 5
	for i := 0; i <= maxRetries; i++ {
		var tableExists bool = doesTableExist(tableName, svc)
		if tableExists && i == maxRetries {
			log.Fatalf("Max retries exceeded! Test table has not been fully deleted, and we cannot proceed.\n")
		} else if tableExists {
			time.Sleep(5 * time.Second)
		} else {
			return true
		}
	}
	return false
}

func hasTableBeenFullyCreated(name string, svc *dynamodb.DynamoDB) bool {
	maxRetries := 5
	for i := 0; i <= maxRetries; i++ {
		var tableExists bool = doesTableExist(tableName, svc)
		if !tableExists && i == maxRetries {
			log.Fatalf("Max retries exceeded! Test table has not been fully created, and we cannot proceed.\n")
		} else if !tableExists {
			time.Sleep(5 * time.Second)
		} else {
			// Check if the table status is "ACTIVE"
			describeTableInput := &dynamodb.DescribeTableInput{
				TableName: aws.String(name),
			}
			tableDescription, err := svc.DescribeTable(describeTableInput)
			if err != nil {
				log.Fatalf("Failed to describe table: %v", err)
				return false
			}
			if *tableDescription.Table.TableStatus == "ACTIVE" {
				return true
			} else {
				// If not yet active, wait and retry
				time.Sleep(5 * time.Second)
			}
		}
	}
	return false
}

// TestMain allows us to do setup and teardown operations before running tests
func TestMain(m *testing.M) {
	/*--------------------------------------------
	SETTING UP THE DB
	--------------------------------------------*/

	AKID := ""
	SECRET_KEY := ""

	// Initialize a session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewStaticCredentials(AKID, SECRET_KEY, ""),
	})
	if err != nil {
		log.Println("Error getting session:")
		log.Fatal(err)
	}
	fmt.Println("After creating test session")

	//create dynamoDB client
	svc = dynamodb.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))

	//check for existence of testTable and initiate deletion it if it exists
	var tableExists bool = doesTableExist(tableName, svc)

	if tableExists {
		deleteTableInput := &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		}

		_, err = svc.DeleteTable(deleteTableInput)
		if err != nil {
			log.Fatalf("Failed to delete test table: %v", err)
			return
		}
		fmt.Printf("Test table %s deleted.\n", tableName)
	}

	//If the testTable has been fully deleted, create a new testTable
	if hasTableBeenFullyDeleted(tableName, svc) == true {
		createTableInput := &dynamodb.CreateTableInput{
			TableName: aws.String(tableName),
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("Creator_Id"), // Partition Key
					KeyType:       aws.String("HASH"),
				},
				{
					AttributeName: aws.String("Creation_Timestamp"), // Sort Key
					KeyType:       aws.String("RANGE"),
				},
			},
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("Creator_Id"),
					AttributeType: aws.String("S"), // S represents String
				},
				{
					AttributeName: aws.String("Creation_Timestamp"),
					AttributeType: aws.String("S"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(5), // Adjust as needed
				WriteCapacityUnits: aws.Int64(5), // Adjust as needed
			},
		}

		_, err = svc.CreateTable(createTableInput)
		if err != nil {
			log.Fatalf("Failed to create test table: %v", err)
			return
		}
		fmt.Printf("Test table %s created.\n", tableName)
	}

	//if the testTable has been fully created, seed the created testTable with data
	if hasTableBeenFullyCreated(tableName, svc) == true {
		var records = testDBSeedData

		for _, record := range records {
			av, err := dynamodbattribute.MarshalMap(record)
			if err != nil {
				log.Fatalf("Failed to marshal record: %v", err)
				return
			}

			putItemInput := &dynamodb.PutItemInput{
				Item:      av,
				TableName: aws.String(tableName),
			}

			_, err = svc.PutItem(putItemInput)
			if err != nil {
				log.Fatalf("Failed to add record: %v", err)
			}
		}
		fmt.Printf("Records added to table %s.\n", tableName)
	}

	/*-------------------------------------------------------
	Initializing a new router and server to use for the tests
	-------------------------------------------------------*/
	testRouter = mux.NewRouter()
	routes.RegisterRoutes(testRouter, svc, tableName)

	testServer = httptest.NewServer(testRouter)
	defer testServer.Close()

	exitCode := m.Run()

	os.Exit(exitCode)
}
