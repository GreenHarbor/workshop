package helpers

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type AttendeesAndVacancy struct {
	Attendees []string
	Vacancies   int
}

func GetAttendeesAndVacancy(creatorID string, creationTimestamp string, svc *dynamodb.DynamoDB, tableName string) (AttendeesAndVacancy, error) {
	var attendeesAndVacancyObject AttendeesAndVacancy
	// Create the QueryInput
	input := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"Creator_Id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(creatorID),
					},
				},
			},
			"Creation_Timestamp": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(creationTimestamp),
					},
				},
			},
		},
	}

	// Perform the Query operation
	result, err := svc.Query(input)
	if err != nil {
		return attendeesAndVacancyObject, err
	} else if *result.Count == int64(0) {
		return attendeesAndVacancyObject, errors.New("Workshop not found.")
	}

	// Unmarshal DynamoDB JSON format to the attendeesAndVacancy model
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &attendeesAndVacancyObject)
	if err != nil {
		return attendeesAndVacancyObject, err
	}
	return attendeesAndVacancyObject, nil
}
