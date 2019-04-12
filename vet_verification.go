package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

//APIError is an error from the VA API
type APIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
	Status string `json:"status"`
}

//VerificationAPIResponse structures data from the VA's Veteran Verification API
//https://developer.va.gov/explore/verification/docs/service_history
type VerificationAPIResponse struct {
	Data []struct {
		ID         string         `json:"id"`
		Type       string         `json:"type"`
		Attributes ServiceHistory `json:"attributes"`
	} `json:"data"`
	Errors []APIError `json:"errors"`
}

//ServiceHistory contains data about a veteran's service history.
type ServiceHistory struct {
	StartDate       string       `json:"start_date"`
	EndDate         string       `json:"end_date"`
	Branch          string       `json:"branch_of_service"`
	DischargeStatus string       `json:"discharge_status"`
	Deployments     []Deployment `json:"deployments"`
}

//Deployment is a veteran's military deployment.
type Deployment struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Location  string `json:"location"`
}

func getServiceHistory(token string) (sh ServiceHistory, err error) {
	req, err := http.NewRequest("GET", baseURL+"/services/veteran_verification/v0/service_history", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Error getting service history:", err)
		return
	}
	defer res.Body.Close()

	var apiResponse VerificationAPIResponse
	err = json.NewDecoder(res.Body).Decode(&apiResponse)
	if err != nil {
		log.Println("Error decoding service history api response:", err)
		return
	}

	if len(apiResponse.Errors) > 0 {
		return sh, errors.New("Error received from Vets API server")
	}

	if len(apiResponse.Data) < 1 {
		return ServiceHistory{}, nil
	}

	return apiResponse.Data[0].Attributes, nil
}
