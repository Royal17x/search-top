package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: go run cmd/stoplist/main.go add|remove <word>")
		os.Exit(1)
	}

	action, word := os.Args[1], os.Args[2]

	switch action {
	case "add":
		body, _ := json.Marshal(map[string]string{"word": word})
		resp, err := http.Post("http://localhost:8080/api/v1/stoplist",
			"application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("status:", resp.Status)

	case "remove":
		req, _ := http.NewRequest(http.MethodDelete,
			"http://localhost:8080/api/v1/stoplist/"+word, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("status:", resp.Status)

	default:
		fmt.Println("unknown action:", action)
		os.Exit(1)
	}
}
