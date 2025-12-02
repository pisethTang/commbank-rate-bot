package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"os"
	"github.com/joho/godotenv"
)

// ------ Configuration ------
const (
	// BotToken   = os.Getenv("TELEGRAM_BOT_TOKEN")
	// ChatID     = os.Getenv("TELEGRAM_CHAT_ID")
	TargetRate = 1.20 // Target USD to AUD rate to trigger notification
)

// Represents a single currency in the list (AUD, GBP, etc.)
type Currency struct {
	Code  string `json:"currencyTitle"` // e.g. "AUD"
	BsImt string `json:"bsImt"`         // e.g. "1.3456"
}

// Represents the entire Json response
type RateReponse struct {
	Currencies []Currency `json:"currencies"`
}

func getRate() (float64, error) {
	// 1. Fetch data from URL with timestamp
	timestamp := time.Now().UnixMilli()
	url := fmt.Sprintf("https://www.commbank.com.au/content/data/forex-rates/USD.json?ts=%d", timestamp)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	// 2. Read body
	body, _ := io.ReadAll(resp.Body)

	// 3. Parse JSON by looping through currencies to find AUD
	var data RateReponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("error parsing JSON: %v", err)
	}

	// 4. Find AUD rate
	for _, currency := range data.Currencies {
		if currency.Code == "AUD" {
			rate, err := strconv.ParseFloat(currency.BsImt, 64)
			if err != nil {
				return 0, fmt.Errorf("error converting rate to float: %v", err)
			}
			return rate, nil
		}
	}
	return 0, fmt.Errorf("AUD rate not found or market closed")
}

func sendTelegram(text string) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		return fmt.Errorf("telegram bot token or chat ID not set in environment variables")
	}

	
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", botToken, chatID, text)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error sending telegram message: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}


	fmt.Println("--- Starting Currency Checker ---")

	rate, err := getRate()
	if err != nil {
		fmt.Printf("Error retrieving rate: %v\n", err)
		return
	}

	// Telegram alert logic
	if rate > TargetRate {
		msg := fmt.Sprintf("🚨 RATE ALERT! 1 USD = %.4f AUD", rate)
		sendTelegram(msg)
		fmt.Println("Alert sent to Telegram!")
	} else {
		fmt.Println("Rate is too low. No alert sent.")
	}

	fmt.Printf("Current USD to AUD rate: %.4f\n", rate)
}
