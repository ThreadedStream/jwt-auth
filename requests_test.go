package main


import(
	"testing"
	"net/http"
	"encoding/json"
	"log"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
)



func makeJSONRequest(client *http.Client, url, method string, params interface{}) *http.Response{
	var jsonStr []byte
	var err error
	bs, ok := params.([]byte)
	if !ok {
		jsonStr, err = json.Marshal(params)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		jsonStr = bs
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
		return nil
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil
	}

	return response

}

func TestObtainTokenPairApi(t *testing.T){
	client := &http.Client{}
	
	url := "http://127.0.0.1:4560/signin"

	correct_params := map[string]interface{}{
		"guid" : "37e3f55c-7c34-439c-ab6d-60644d23cc7f",
	}

	fake_params := map[string]interface{}{
		"guid": "23213221312-21321312312",
	}

	//This one is supposed to succeed
	responseOK := makeJSONRequest(client, url, "POST", correct_params)

	defer responseOK.Body.Close()

	assert.NotEqual(t, responseOK, nil)
	assert.Equal(t, responseOK.StatusCode, http.StatusOK)

	//That one is supposed to fail
	responseBR := makeJSONRequest(client, url, "POST", fake_params)

	defer responseBR.Body.Close()

	assert.NotEqual(t, responseBR, nil)
	assert.Equal(t, responseBR.StatusCode, http.StatusBadRequest)
}

func TestRefreshApi(t *testing.T){
	client := &http.Client{}
	
	url := "http://127.0.0.1:4560/refresh"

	params := map[string]interface{}{
		"access_token" : "wadawdawda",
		"refresh_token": "wwadawdawdawdawdawdawd",
	}	

	response := makeJSONRequest(client, url, "POST", params)

	assert.NotEqual(t, response, nil)
	assert.Equal(t, response.StatusCode, http.StatusUnauthorized)
}

