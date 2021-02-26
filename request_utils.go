package main

import (
	"encoding/json"
	"net/http"
)

type RequestEntities struct {
	Guid         string `json:"guid"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type Response struct {
	w               http.ResponseWriter
	statusCode      int
	responseMessage interface{}
}

func (r *Response) JsonResponse() {
	response, _ := json.Marshal(r.responseMessage)

	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.statusCode)
	r.w.Write(response)
}

func ResponseShortcut(w http.ResponseWriter, statusCode int, responseMessage interface{}) {
	response := Response{
		w:               w,
		statusCode:      statusCode,
		responseMessage: responseMessage,
	}
	response.JsonResponse()
}
