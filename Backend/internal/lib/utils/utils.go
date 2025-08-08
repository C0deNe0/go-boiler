package utils

import (
	"encoding/json"
	"fmt"
)

func PrintJSON(v interface{}) {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error Marshalling to JSON:", err)
		return
	}
	fmt.Println("JSON:", string(json))
}
