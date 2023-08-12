package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	db, _ = sql.Open("sqlite3", "./cotacoes.db")
	defer db.Close()

	http.HandleFunc("/cotacao", handleCotacao)
	log.Fatal(http.ListenAndServe(":9080", nil))
}

func handleCotacao(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go saveCotacaoToDB(ctx)

	client := http.Client{}
	req, err := http.NewRequest("GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req = req.WithContext(ctx)
	fmt.Println(req)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error making request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	cotacao, ok := result["USDBRL"].(map[string]interface{})
	if !ok {
		http.Error(w, "Error getting cotacao value from response", http.StatusInternalServerError)
		return
	}

	cotacaoValue, ok := cotacao["bid"].(string)
	if !ok {
		http.Error(w, "Error getting bid value from response", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"bid": cotacaoValue}
	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func saveCotacaoToDB(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Saving to DB cancelled")
		return
	default:
		time.Sleep(10 * time.Millisecond)

		cotacaoValue := "example_cotacao_value"
		_, err := db.Exec("INSERT INTO cotacoes (valor) VALUES (?)", cotacaoValue)
		if err != nil {
			log.Println("Error saving to DB:", err)
		} else {
			log.Println("Cotacao saved to DB")
		}
	}
}
