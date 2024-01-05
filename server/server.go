package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
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
		log.Printf("[Error] Get quotation error: %v\n", err)
		return
	}

	dbSqlite, err := OpenDatabase()
	if err != nil {
		log.Printf("[Error] Open database error: %v\n", err)
		return
	}

	err = SaveQuotation(dbSqlite, data.UsdBrl)
	if err != nil {
		log.Printf("[Error] Salve quotation error: %v\n", err)
		return
	}

	res, err := json.Marshal(data.UsdBrl)
	if err != nil {
		log.Printf("[Error] Parse json error: %v\n", err)
		return
	}

	_, err = w.Write(res)
	if err != nil {
		log.Printf("[Error] Write response error: %v\n", err)
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

func SaveQuotation(dbSqlite *gorm.DB, dto QuotationDto) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	tx := dbSqlite.WithContext(ctx)
	var c = NewQuotationEntity(dto)
	err := tx.Save(&c).Error
	if err != nil {
		return err
	}
	return nil
}
