package server

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func (s *FiberServer) DefaultOhlcvHandler(c *fiber.Ctx) error {
	println("DefaultOhlcvHandler match")
	token := "BTC"
	tokenAddresses := map[string]string{
		"BTC": "TN3W4H6rK2ce4vX9YnFQHwKENnHjoxb3m9",
		"ETH": "0x2170Ed0880ac9A755fd29B268",
	}

	address, exists := tokenAddresses[token]
	if !exists {
		return c.Status(400).SendString("Invalid token")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.syve.ai/v1/price/historical/ohlc?token_address=%s&price_type=price_token_usd_tick_1&pool_address=all", address)
	resp, err := client.Get(url)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(500).SendString(fmt.Sprintf("API responded with status code: %d", resp.StatusCode))
	}

	// Parse the response JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Extract and reshape the data
	data, ok := result["data"].([]interface{})
	if !ok {
		return c.Status(500).SendString("Invalid data format")
	}

	formattedData := make([][]interface{}, len(data))
	for i, entry := range data {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			return c.Status(500).SendString("Invalid data format")
		}

		timestamp, ok := entryMap["timestamp_open"].(float64)
		if !ok {
			return c.Status(500).SendString("Invalid timestamp format")
		}

		ohlc := []float64{
			entryMap["price_open"].(float64),
			entryMap["price_high"].(float64),
			entryMap["price_low"].(float64),
			entryMap["price_close"].(float64),
		}

		formattedData[i] = []interface{}{int64(timestamp), ohlc}
	}

	// Return the formatted data as JSON response
	return c.JSON(fiber.Map{
		"data": formattedData,
	})
}

func (s *FiberServer) OhlcvHandler(c *fiber.Ctx) error {
	println("OhlcvHandler match")
	tokenAddresses := map[string]string{
		"BTC": "TN3W4H6rK2ce4vX9YnFQHwKENnHjoxb3m9",
		"ETH": "0x2170Ed0880ac9A755fd29B268",
	}

	token := strings.ToUpper(c.Params("token"))
	address, exists := tokenAddresses[token]
	if !exists {
		return c.Status(400).SendString("Invalid token")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.syve.ai/v1/price/historical/ohlc?token_address=%s&price_type=price_token_usd_tick_1&pool_address=all", address)
	resp, err := client.Get(url)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(500).SendString(fmt.Sprintf("API responded with status code: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		return c.Status(500).SendString("Invalid data format")
	}

	formattedData := make([][]interface{}, len(data))
	for i, entry := range data {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			return c.Status(500).SendString("Invalid data format")
		}

		timestamp, ok := entryMap["timestamp_open"].(float64)
		if !ok {
			return c.Status(500).SendString("Invalid timestamp format")
		}

		ohlc := []float64{
			entryMap["price_open"].(float64),
			entryMap["price_high"].(float64),
			entryMap["price_low"].(float64),
			entryMap["price_close"].(float64),
		}

		formattedData[i] = []interface{}{int64(timestamp), ohlc}
	}

	return c.JSON(fiber.Map{
		"data": formattedData,
	})
}
