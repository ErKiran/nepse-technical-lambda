package main

import (
	"context"
	"encoding/json"
	"fmt"
	"nepse-technical-gateway-lambda/nepse"
	"nepse-technical-gateway-lambda/utils"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

type Ticker struct {
	Symbol string
}

type TickerResponse struct {
	RSI        map[string][]float64 `json:"rsi"`
	MACD       map[string][]float64 `json:"macd"`
	SignalLine map[string][]float64 `json:"signalLine"`
	Histogram  map[string][]float64 `json:"histogram"`
	Ema20      map[string][]float64 `json:"ema20"`
	Ema50      map[string][]float64 `json:"ema50"`
	Ema200     map[string][]float64 `json:"ema200"`
	KeyLevels  utils.KeyLevels      `json:"keyLevels"`
}

func TechnicalHandler(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("err", err)
		os.Exit(0)
	}

	rsiMap := make(map[string][]float64)

	macdMap := make(map[string][]float64)
	signalLineMap := make(map[string][]float64)
	histogramMap := make(map[string][]float64)

	ema20Map := make(map[string][]float64)
	ema50Map := make(map[string][]float64)
	ema200Map := make(map[string][]float64)

	var keyLevels utils.KeyLevels

	nepse, err := nepse.NewNepse()

	if err != nil {
		return nil, err
	}

	var tickers = []Ticker{
		{Symbol: "NABIL"},
		{Symbol: "MNBBL"},
	}

	for _, stock := range tickers {

		data, err := nepse.GetTechnicalData(stock.Symbol, "D")
		if err != nil {
			return nil, err
		}

		rsiMap[stock.Symbol] = data.RSI()
		macdMap[stock.Symbol], signalLineMap[stock.Symbol], histogramMap[stock.Symbol] = data.MACD()
		ema20Map[stock.Symbol] = data.EMA(20)
		ema50Map[stock.Symbol] = data.EMA(50)
		ema200Map[stock.Symbol] = data.EMA(200)
		keyLevels = data.KeyLevels()
	}

	var response = TickerResponse{
		RSI:        rsiMap,
		MACD:       macdMap,
		SignalLine: signalLineMap,
		Histogram:  histogramMap,
		Ema20:      ema20Map,
		Ema50:      ema50Map,
		Ema200:     ema200Map,
		KeyLevels:  keyLevels,
	}

	b, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))

	return &events.APIGatewayProxyResponse{
		Body:       string(b),
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(TechnicalHandler)
}
