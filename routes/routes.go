package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"workshop/helpers"
	"workshop/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	"github.com/thoas/go-funk"
)
var tableName string

func RegisterRoutes(r *mux.Router, svc *dynamodb.DynamoDB, table string) {
	tableName = table
	r.HandleFunc("/health", health_check)
	r.HandleFunc("/workshop", get_all(svc)).Methods("GET")
	r.HandleFunc("/workshop/{creator_id}", get_by_creatorID(svc)).Methods("GET")
	r.HandleFunc("/workshop", create(svc)).Methods("POST")
	r.HandleFunc("/workshop/{creator_id}/{creation_timestamp}", patch(svc)).Methods("PATCH")
	r.HandleFunc("/workshop/{creator_id}/{creation_timestamp}", delete(svc)).Methods("DELETE")
	r.HandleFunc("/workshop/register/{creator_id}/{creation_timestamp}", register(svc)).Methods("PATCH")
	r.HandleFunc("/workshop/withdraw/{creator_id}/{creation_timestamp}", withdraw(svc)).Methods("PATCH")
}

func health_check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := make(map[string]string)
	resp["message"] = "Service is healthy"
	resp["service"] = "Workshop"
	jsonResp, err := json.Marshal(resp)

	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		w.WriteHeader(500)
		return
	}

	if _, err := w.Write(jsonResp); err != nil {
		log.Fatalf("Unable to write JSON: %s", err)
		return
	}

	w.WriteHeader(200)
}

func get_all(svc *dynamodb.DynamoDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)
			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}
	
		// Build the dynamoDB scan input
		input := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}
		// Perform the scan on dynamoDB
		result, err := svc.Scan(input)
		if err != nil {
			errMsg := err.Error()
			handleError(errMsg, 500)
			return
		}
	
		var workshops []interface{}
	
		for _, i := range result.Items {
			// Unmarshal DynamoDB JSON format to the workshop model
			var workshop_model models.Workshop
			err = dynamodbattribute.UnmarshalMap(i, &workshop_model)
			if err != nil {
				handleError("Error unmarshalling dynamoDB JSON format to the workshop model", 500)
				return
			}
	
			// Append the unmarshalled item to the workshops array
			workshops = append(workshops, workshop_model)
		}
	
		// Marshal the workshops array to create a JSON array
		workshopsJSON, err := json.Marshal(workshops)
		if err != nil {
			handleError("Error marshalling workshop models to JSON", 500)
		}
		// Send the JSON response to the client
		w.WriteHeader(http.StatusOK)
	
		if _, err := w.Write(workshopsJSON); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func get_by_creatorID(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)
			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}

		//get the partition key, creator_id from the url
		vars := mux.Vars(r)
		creatorID := vars["creator_id"]

		// Create the QueryInput
		input := &dynamodb.QueryInput{
			TableName:              aws.String(tableName),
			KeyConditionExpression: aws.String("Creator_Id = :val"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":val": {
					S: aws.String(creatorID),
				},
			},
		}
		// Perform the Query operation
		result, err := svc.Query(input)
		if err != nil {
			errorMsg := "Error querying items with Creator_Id: " + creatorID
			handleError(errorMsg, 404)
			return
		}

		var workshops []interface{}
		for _, i := range result.Items {
			// Unmarshal DynamoDB JSON format to the workshop model
			var workshop_model models.Workshop
			err = dynamodbattribute.UnmarshalMap(i, &workshop_model)
			if err != nil {
				handleError("Error unmarshalling dynamoDB JSON format to the workshop model", 500)
				return
			}
			// Append the unmarshalled item to the workshops array
			workshops = append(workshops, workshop_model)
		}

		// Marshal the workshops array to create a JSON array
		workshopsJSON, err := json.Marshal(workshops)
		if err != nil {
			handleError("Error marshalling workshop models to JSON", 500)
		}
		// Send the JSON response to the client
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(workshopsJSON); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func create(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)

			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}

		// Parse the request body into the Workshop struct
		var request models.Workshop
		err := json.NewDecoder(r.Body).Decode(&request)
		if request.Creator_Id == "" {
			handleError("Missing creator_ID", 400)
			return
		}
		if err != nil {
			handleError("Invalid request data.", 400)
			return
		}
		//append a creation timestamp and empty attendees list to the request body
		currentTimeUTC := time.Now().UTC().Add(8 * time.Hour)
		request.Creation_Timestamp = currentTimeUTC.Format("2006-01-02-15:04:05.000")
		request.Attendees = []string{}

		//marshall the struct into an attribute value object
		av, err := dynamodbattribute.MarshalMap(request)
		if err != nil {
			handleError("Error marshalling data into an attribute value object.", 400)
			return
		}
		// Insert the data into the database
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}
		_, err = svc.PutItem(input)
		if err != nil {
			handleError("Error inserting workshop data into the database.", 500)
			return
		}

		resp["message"] = "Workshop created successfully."
		w.WriteHeader(201)
		jsonResponse, _ := json.Marshal(resp)

		if _, err := w.Write(jsonResponse); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func patch(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)
			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}

		//get the partition and sort key from the url and parse it into key
		vars := mux.Vars(r)
		creatorID := vars["creator_id"]
		creationTimestamp := vars["creation_timestamp"]
		key := map[string]*dynamodb.AttributeValue{
			"Creator_Id": {
				S: aws.String(creatorID),
			},
			"Creation_Timestamp": {
				S: aws.String(creationTimestamp),
			},
		}

		// Create a map to hold the fields from the JSON request body
		var updateFields map[string]interface{}
		// Parse the JSON request body into the updateFields map
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&updateFields); err != nil {
			handleError("Invalid Request Data.", 400)
			return
		}

		// create the dynamoDB update expression and map to hold the expression attribute values
		updateExpression := "SET "
		expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
		//loop through updatefields map to populate the updateExpression and expressionAttributeValues
		for key, value := range updateFields {
			attrValue := &dynamodb.AttributeValue{}

			switch v := value.(type) {
			case string:
				attrValue.S = aws.String(v)
			case float64: //unmarshalled json ints are often converted to float64
				intValue := int(v)
				attrValue.N = aws.String(strconv.Itoa(intValue))
			case int:
				attrValue.N = aws.String(strconv.Itoa(v))
			default:
				handleError("You may not patch this field", 400)
				return
			}
			expressionAttributeValues[":"+key] = attrValue
			updateExpression = updateExpression + key + " = :" + key + ", "
		}
		updateExpression = updateExpression[:len(updateExpression)-2]

		// Specify the update input.
		updateInput := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       key,
			UpdateExpression:          aws.String(updateExpression),
			ExpressionAttributeValues: expressionAttributeValues,
		}

		// Execute the update operation.
		_, err := svc.UpdateItem(updateInput)
		if err != nil {
			handleError("Error updating the database", 500)
			return
		}
		resp["message"] = "Workshop updated successfully."
		w.WriteHeader(200)
		jsonResponse, _ := json.Marshal(resp)

		if _, err := w.Write(jsonResponse); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func delete(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)

			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}

		//get the partition and sort key from the url and parse it into key
		vars := mux.Vars(r)
		creatorID := vars["creator_id"]
		creationTimestamp := vars["creation_timestamp"]
		key := map[string]*dynamodb.AttributeValue{
			"Creator_Id": {
				S: aws.String(creatorID),
			},
			"Creation_Timestamp": {
				S: aws.String(creationTimestamp),
			},
		}

		// Define the input for the DeleteItem operation.
		input := &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key:       key,
		}
		// Delete the item.
		_, err := svc.DeleteItem(input)
		if err != nil {
			handleError("Unable to delete item. Check if creatorID and creationTimestamp is correct?", 500)
			return
		}

		resp["message"] = fmt.Sprintf("Workshop with creator_id %s and creation_timestamp %s deleted successfully.", creatorID, creationTimestamp)
		w.WriteHeader(200)
		jsonResponse, _ := json.Marshal(resp)

		if _, err := w.Write(jsonResponse); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func register(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)

			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}

		//get the partition and sort key from the url and parse it into key
		vars := mux.Vars(r)
		creatorID := vars["creator_id"]
		creationTimestamp := vars["creation_timestamp"]
		key := map[string]*dynamodb.AttributeValue{
			"Creator_Id": {
				S: aws.String(creatorID),
			},
			"Creation_Timestamp": {
				S: aws.String(creationTimestamp),
			},
		}
		// Extract userID from the JSON request body
		var requestBody map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			handleError("Invalid Request Data.", 400)
			return
		}
		var userID string
		if user_id, ok := requestBody["User_Id"].(string); ok {
			userID = user_id
		} else {
			handleError("User_Id given is not a string!", 400)
			return
		}
		//get the attendees and vacancy of the current workshop
		var attendeesAndVacancy helpers.AttendeesAndVacancy
		attendeesAndVacancy, err := helpers.GetAttendeesAndVacancy(creatorID, creationTimestamp, svc, tableName)
		if err != nil {
			errorMessage := err.Error()
			if errorMessage == "Workshop not found." {
				handleError(errorMessage, 404)
			} else {
				handleError(errorMessage, 500)
			}
			return
		}
		attendees := attendeesAndVacancy.Attendees
		vacancies := attendeesAndVacancy.Vacancies
		//IF USER IS ALREADY IN ATTENDEE LIST, return an error
		if funk.Contains(attendees, userID) {
			handleError("User is already in attendees list!", 400)
			return
		}
		//IF THERE IS VACANCY, add the user to attendee list, and decrement vacancy
		if vacancies == 0 {
			handleError("There is 0 vacancy!", 500)
			return
		} else {
			attendees = append(attendees, userID)
			vacancies -= 1
		}
		// Convert the list of attendees to a list of DynamoDB AttributeValues
		attendeesAttributeValues := make([]*dynamodb.AttributeValue, len(attendees))
		for i, uid := range attendees {
			attendeesAttributeValues[i] = &dynamodb.AttributeValue{
				S: aws.String(uid),
			}
		}
		// Define the update expression and attribute values to send to dynamoDB, the buggering database who designed this
		updateExpression := "SET Attendees = :value1, Vacancies = :value2"
		expressionAttributeValues := map[string]*dynamodb.AttributeValue{
			":value1": {
				L: attendeesAttributeValues,
			},
			":value2": {
				N: aws.String(strconv.Itoa(vacancies)),
			},
		}
		// Specify the update input.
		updateInput := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       key,
			UpdateExpression:          aws.String(updateExpression),
			ExpressionAttributeValues: expressionAttributeValues,
		}
		// Execute the update operation.
		_, err = svc.UpdateItem(updateInput)
		if err != nil {
			handleError("Error updating the database", 500)
			return
		}
		resp["message"] = "Registration successful!"
		w.WriteHeader(200)
		jsonResponse, _ := json.Marshal(resp)

		if _, err := w.Write(jsonResponse); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}

func withdraw(svc *dynamodb.DynamoDB) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		handleError := func(message string, statusCode int) {
			resp["message"] = message
			w.WriteHeader(statusCode)
			jsonResponse, _ := json.Marshal(resp)

			if _, err := w.Write(jsonResponse); err != nil {
				log.Fatalf("Unable to write JSON: %s", err)
				return
			}
		}
		//get the partition and sort key from the url and parse it into key
		vars := mux.Vars(r)
		creatorID := vars["creator_id"]
		creationTimestamp := vars["creation_timestamp"]
		key := map[string]*dynamodb.AttributeValue{
			"Creator_Id": {
				S: aws.String(creatorID),
			},
			"Creation_Timestamp": {
				S: aws.String(creationTimestamp),
			},
		}
		// Extract userID from the JSON request body
		var requestBody map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			handleError("Invalid Request Data.", 400)
			return
		}
		var userID string
		if user_id, ok := requestBody["User_Id"].(string); ok {
			userID = user_id
		} else {
			handleError("User_Id given is not a string!", 400)
			return
		}
		//get the attendees and vacancy of the current workshop
		var attendeesAndVacancy helpers.AttendeesAndVacancy
		attendeesAndVacancy, err := helpers.GetAttendeesAndVacancy(creatorID, creationTimestamp, svc, tableName)
		if err != nil {
			errorMessage := err.Error()
			if errorMessage == "Workshop not found." {
				handleError(errorMessage, 404)
			} else {
				handleError(errorMessage, 500)
			}
			return
		}
		//IF userID is in the attendees list, remove userID from attendee list, and increment vacancy
		attendees := attendeesAndVacancy.Attendees
		vacancies := attendeesAndVacancy.Vacancies
		if funk.Contains(attendees, userID) {
			index := funk.IndexOf(attendees, userID)
			newAttendees, err := helpers.RemoveFromList(attendees, index)
			if err != nil {
				handleError(err.Error(), 400)
				return
			}
			attendees = newAttendees
			vacancies += 1
		} else {
			handleError("UserID not found in the attendees list!", 400)
			return
		}

		// Convert the list of attendees to a list of DynamoDB AttributeValues
		attendeesAttributeValues := make([]*dynamodb.AttributeValue, len(attendees))
		for i, uid := range attendees {
			attendeesAttributeValues[i] = &dynamodb.AttributeValue{
				S: aws.String(uid),
			}
		}
		// Define the update expression and attribute values to send to dynamoDB, the buggering database who designed this
		updateExpression := "SET Attendees = :value1, Vacancies = :value2"
		expressionAttributeValues := map[string]*dynamodb.AttributeValue{
			":value1": {
				L: attendeesAttributeValues,
			},
			":value2": {
				N: aws.String(strconv.Itoa(vacancies)),
			},
		}
		// Specify the update input.
		updateInput := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       key,
			UpdateExpression:          aws.String(updateExpression),
			ExpressionAttributeValues: expressionAttributeValues,
		}
		// Execute the update operation.
		_, err = svc.UpdateItem(updateInput)
		if err != nil {
			handleError("Error updating the database", 500)
			return
		}
		resp["message"] = "Withdrawal successful!"
		w.WriteHeader(200)
		jsonResponse, _ := json.Marshal(resp)

		if _, err := w.Write(jsonResponse); err != nil {
			log.Fatalf("Unable to write JSON: %s", err)
			return
		}
	}
}