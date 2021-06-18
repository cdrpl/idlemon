package main

import (
	"encoding/json"
)

type Resource struct {
	Name string `json:"name"`
}

type UserResource struct {
	Amount int `json:"amount"`
}

// Unmarshall the embeded resourcesJson string.
func UnmarshallResourcesJson() ([]Resource, error) {
	var data map[string][]Resource

	err := json.Unmarshal([]byte(resourcesJson), &data)
	if err != nil {
		return nil, err
	}

	return data["resources"], nil
}
