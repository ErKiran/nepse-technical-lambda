package nepse

import (
	"context"
	"fmt"
	"nepse-technical-gateway-lambda/utils"
	"net/http"
	"os"
	"time"
)

const (
	Health    = "overview/topGainers/?count=5"
	Technical = "tradingView/history"
)

type Nepse struct {
	client *utils.Client
}

func NewNepse() (*Nepse, error) {
	client := utils.NewClient(nil, os.Getenv("NEPSE"))

	_, err := client.NewRequest(http.MethodGet, Health, nil)

	if err != nil {
		return nil, err
	}

	nep := &Nepse{
		client: client,
	}
	return nep, nil
}

func (n Nepse) buildTickerSlugTechnicalURL(urlPath, ticker, resoultion string, start, end int64) string {
	return fmt.Sprintf("%s?symbol=%s&resolution=%s&from=%d&to=%d", urlPath, ticker, resoultion, start, end)
}

func (n Nepse) GetTechnicalData(stock, resolution string) (*utils.TechnicalData, error) {
	now := time.Now()
	start := now.AddDate(-1, 0, -5)
	url := n.buildTickerSlugTechnicalURL(Technical, stock, resolution, start.Unix(), now.Unix())
	fmt.Println("url", url)
	req, err := n.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res := &utils.TechnicalData{}
	if _, err := n.client.Do(context.Background(), req, res); err != nil {
		return nil, err
	}

	return res, nil
}
