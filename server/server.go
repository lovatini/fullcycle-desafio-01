package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type QuotationUsdBrlDto struct {
	UsdBrl QuotationDto `json:"USDBRL"`
}

type QuotationDto struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type QuotationEntity struct {
	ID         string `gorm:"primaryKey"`
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

func NewQuotationEntity(dto QuotationDto) *QuotationEntity {
	return &QuotationEntity{
		ID:         uuid.New().String(),
		Code:       dto.Code,
		Codein:     dto.Codein,
		Name:       dto.Name,
		High:       dto.High,
		Low:        dto.Low,
		VarBid:     dto.VarBid,
		PctChange:  dto.PctChange,
		Bid:        dto.Bid,
		Ask:        dto.Ask,
		Timestamp:  dto.Timestamp,
		CreateDate: dto.CreateDate,
	}
}

func main() {
	http.HandleFunc("/cotacao", QuotationHandler)
	log.Println("Start => http://localhost:9000/cotacao")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func QuotationHandler(w http.ResponseWriter, r *http.Request) {
	var data, err = GetQuotationClient()
	if err != nil {
		LogError(err)
		return
	}

	dbSqlite, err := OpenDatabase()
	if err != nil {
		LogError(err)
		return
	}
	SaveQuotation(dbSqlite, data.UsdBrl)

	res, err := json.Marshal(data.UsdBrl)
	if err != nil {
		LogError(err)
		return
	}

	_, err = w.Write(res)
	if err != nil {
		LogError(err)
		return
	}
}

func GetQuotationClient() (QuotationUsdBrlDto, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	requestHttp, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://economia.awesomeapi.com.br/json/last/USD-BRL",
		nil,
	)
	if err != nil {
		return QuotationUsdBrlDto{}, err
	}

	resultHttp, err := http.DefaultClient.Do(requestHttp)
	if err != nil {
		return QuotationUsdBrlDto{}, err
	}
	defer resultHttp.Body.Close()

	readHttp, err := io.ReadAll(resultHttp.Body)
	if err != nil {
		return QuotationUsdBrlDto{}, err
	}

	var data QuotationUsdBrlDto
	err = json.Unmarshal(readHttp, &data)
	if err != nil {
		return QuotationUsdBrlDto{}, err
	}
	return data, nil
}

func OpenDatabase() (*gorm.DB, error) {
	dbSqlite, err := gorm.Open(sqlite.Open("fullcycle_desafio_01.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dbSqlite.AutoMigrate(&QuotationEntity{})
	if err != nil {
		return nil, err
	}
	return dbSqlite, err
}

func SaveQuotation(dbSqlite *gorm.DB, dto QuotationDto) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	tx := dbSqlite.WithContext(ctx).Begin()
	var c = NewQuotationEntity(dto)
	tx.Debug().Save(&c)
	tx.Commit()
}

func LogError(errLog error) {
	log.Printf("Error: %v\n\n", errLog)
	_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", errLog)
	if err != nil {
		return
	}
}
