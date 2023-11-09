package tests

import(
	"encoding/json"
	"net/http"
	"testing"
	"log"
	"bytes"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
    res, _ := http.Get(testServer.URL + "/health")
    assert.Equal(t, 200, res.StatusCode, "Expected result to be %d, but got %d", 200, res.StatusCode)
}

func TestGetAll( t *testing.T) {
	res, err := http.Get(testServer.URL + "/workshop")
    if err != nil {
        log.Fatalf("Failed to send the HTTP request: %v", err)
    }
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Fatalf("Failed to read the response body in TestGetAll: %v", err)
    }
	//Convert data that we tried to seed into the DB to string representation
    testDBSeedDataJson, err := json.Marshal(testDBSeedData)
    if err != nil {
        log.Fatalf("Failed to marshal testDBSeedData to JSON in TestGetAll: %v", err)
    }
	testDBSeedDataJsonString := string(testDBSeedDataJson)

    //check if we get what is expected
    assert.Equal(t, 200, res.StatusCode, "Expected result to be %d, but got %d", 200, res.StatusCode)
	assert.Equal(t, testDBSeedDataJsonString, string(body), "Expected result to be %s, but got %s", testDBSeedDataJsonString, string(body))
}

func TestCreate(t *testing.T) {
	requestBody := map[string]interface{}{
		"Creator_Id": "2",
		"Title": "TEST WORKSHOP!",
		"Description": "This is for testing", 
		"Location": "123 Test Road",
		"Vacancies": 22,
		"Registration_Deadline": "2024-02-17-23:59:59.000",
		"Start_Timestamp": "2024-02-10-15:00:00.000",
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Failed to marshal requestBody JSON in TestCreate: %v", err)
	}
		
	res, err := http.Post(testServer.URL + "/workshop", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("TestCreate has failed-- post request could not go through: %v", err)
	}
    assert.Equal(t, 201, res.StatusCode, "Expected result to be %d, but got %d", 201, res.StatusCode)
}

// func TestGetByCreatorID(t *testing.T) {

// }


// func TestPatch(t *testing.T) {
	
// }

// func TestDelete(t *testing.T) {

// }

// func TestRegister(t *testing.T) {

// }

// func TestWithdraw(t *testing.T) {

// }