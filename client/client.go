package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ResulDto struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	requestHttp, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9000/cotacao", nil)
	if err != nil {
		LogError("Error reading resquest http: %v", err)
		return
	}

	responseHttp, err := http.DefaultClient.Do(requestHttp)
	if err != nil {
		LogError("Error reading response http: %v", err)
		return
	}
	defer responseHttp.Body.Close()

	resultInJson, err := io.ReadAll(responseHttp.Body)
	if err != nil {
		LogError("Error reading response http in bytes: %v", err)
		return
	}

	var data ResulDto
	err = json.Unmarshal(resultInJson, &data)
	if err != nil {
		LogError("Error parser json: %v", err)
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		LogError("Create file: %v", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar:  %s", data.Bid))
	fmt.Println("File created successfully!")
	fmt.Println("Dólar: ", data.Bid)
}

func LogError(mensagem string, errLog error) {
	log.Printf("Error: %v\n\n", errLog)
	_, err := fmt.Fprintf(os.Stderr, mensagem, errLog)
	if err != nil {
		return
	}
}
