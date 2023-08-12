package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	client := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:9080/cotacao", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	cotacao, ok := result["bid"].(string)
	if !ok {
		fmt.Println("Error getting bid value from response")
		return
	}

	fmt.Println("Cotação atual:", cotacao)

	fileContent := fmt.Sprintf("Dólar: %s\n", cotacao)
	err = os.WriteFile("cotacao.txt", []byte(fileContent), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
