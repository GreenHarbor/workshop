package helpers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSCredentials struct {
	AWSAccessKeyID     string 
	AWSSecretAccessKey string 
}

func GetAccessKeys() (AWSCredentials, error) {
	var creds AWSCredentials

	secretName := "arn:aws:secretsmanager:ap-southeast-1:650654341870:secret:workshopAccessToDynamoDB-8hAXae"
	region := "ap-southeast-1"

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("Error in loading default config")
		log.Println(err)
		return creds, err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Println("Error in getting secret value")
		log.Println(err)
		return creds, err
	}

	var secretString string = *result.SecretString
	log.Println(secretString)

	// Unmarshal the JSON string into the struct
	if err := json.Unmarshal([]byte(secretString), &creds); err != nil {
		return creds, err
	}
	log.Println(creds)
	return creds, nil
}
