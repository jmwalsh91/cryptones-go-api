package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/", s.HelloWorldHandler)
	s.App.Get("api/ohlcv", s.DefaultOhlcvHandler)
	s.App.Get("/api/ohlcv/:token/:interval", s.OhlcvHandler)
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}
	return c.JSON(resp)
}

type OHLCVData struct {
	TimePeriodStart string  `json:"time_period_start"`
	TimePeriodEnd   string  `json:"time_period_end"`
	TimeOpen        string  `json:"time_open"`
	TimeClose       string  `json:"time_close"`
	PriceOpen       float64 `json:"price_open"`
	PriceHigh       float64 `json:"price_high"`
	PriceLow        float64 `json:"price_low"`
	PriceClose      float64 `json:"price_close"`
	VolumeTraded    float64 `json:"volume_traded"`
	TradesCount     int     `json:"trades_count"`
}

func (s *FiberServer) DefaultOhlcvHandler(c *fiber.Ctx) error {
	apiKey := os.Getenv("API_KEY")
	url := "https://rest.coinapi.io/v1/ohlcv/BITSTAMP_SPOT_BTC_USD/latest?period_id=5MIN"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	req.Header.Set("X-CoinAPI-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(500).SendString(fmt.Sprintf("API responded with status code: %d", resp.StatusCode))
	}

	var ohlcvData []OHLCVData
	if err := json.NewDecoder(resp.Body).Decode(&ohlcvData); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	resFormatted := make([][]interface{}, len(ohlcvData))
	volArr := make([]float64, len(ohlcvData))
	for i, data := range ohlcvData {
		resFormatted[i] = []interface{}{
			data.TimePeriodStart,
			[]float64{
				data.PriceOpen,
				data.PriceHigh,
				data.PriceLow,
				data.PriceClose,
			},
		}
		volArr[i] = data.VolumeTraded
	}

	return c.JSON(fiber.Map{
		"formattedOhlc": resFormatted,
		"volArr":        volArr,
		"tokenName":     "BTC",
		"interval":      "5MIN",
	})
}

func (s *FiberServer) OhlcvHandler(c *fiber.Ctx) error {
	token := strings.ToUpper(c.Params("token"))
	interval := c.Params("interval")

	symbolID := ""
	switch token {
	case "BTC":
		symbolID = "BITSTAMP_SPOT_BTC_USD"
	case "ETH":
		symbolID = "BITSTAMP_SPOT_ETH_USD"
	default:
		return c.Status(400).SendString("Invalid token")
	}

	apiKey := os.Getenv("API_KEY")
	url := fmt.Sprintf("https://rest.coinapi.io/v1/ohlcv/%s/latest?period_id=%s", symbolID, interval)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	req.Header.Set("X-CoinAPI-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(500).SendString(fmt.Sprintf("API responded with status code: %d", resp.StatusCode))
	}

	var ohlcvData []OHLCVData
	if err := json.NewDecoder(resp.Body).Decode(&ohlcvData); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	resFormatted := make([][]interface{}, len(ohlcvData))
	volArr := make([]float64, len(ohlcvData))
	for i, data := range ohlcvData {
		resFormatted[i] = []interface{}{
			data.TimePeriodStart,
			[]float64{
				data.PriceOpen,
				data.PriceHigh,
				data.PriceLow,
				data.PriceClose,
			},
		}
		volArr[i] = data.VolumeTraded
	}

	return c.JSON(fiber.Map{
		"formattedOhlc": resFormatted,
		"volArr":        volArr,
		"tokenName":     token,
		"interval":      interval,
	})
}
