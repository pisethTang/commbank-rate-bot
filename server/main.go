package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"conversion-bot/config"

	"github.com/joho/godotenv"
)

// global mutex to protect rate access
var (
	currentRate float64
	lastUpdated time.Time
	mu          sync.RWMutex // mutex to protect access to currentRate and lastUpload
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
		// fmt.Println("Error loading .env file")
		fmt.Println("Warning: .env file not found. Using system environment variables.") // return
	}

	cfg := config.Initialize()

	fmt.Printf("--- Starting Currency Checker Bot (%v) ---\n", cfg.CheckInterval)

	// run check immediately, then at intervals
	checkAndNotify(cfg) // Initial check

	//  starts the ticker (periodic checks at intervals) in a BACKGROUND thread using go routine
	go func() {
		ticker := time.NewTicker(cfg.CheckInterval)
		// defer ticker.Stop()

		// this loop runs at each tick
		for range ticker.C {
			checkAndNotify(cfg)
		}
	}()

	// health check endpoint
	http.HandleFunc("/health", handleHealth)
	// start the web server in a main thread
	http.HandleFunc("/api/rate", handleGetRate)
	// 4. THE BLOCKING CALL (must be last)
	fmt.Println("Server started at 0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		// Use log.Fatal to ensure the exit code isn't 0 if it fails
		fmt.Printf("CRITICAL: Server failed: %v\n", err)
		os.Exit(1)
	}

}

func checkAndNotify(cfg config.Config) {
	timestamp := time.Now().Format("15:04:05") // for logging
	fmt.Printf("[%s] Checking rates ...", timestamp)

	rate, err := getRate()
	if err != nil {
		fmt.Printf(" Error retrieving rate: %v\n", err)
		return
	}

	// update global rate and timestamp with mutex
	mu.Lock()
	currentRate = rate
	lastUpdated = time.Now()
	mu.Unlock()

	fmt.Printf(" Current USD to AUD rate: %.4f\n", rate)

	// Telegram alert logic
	if rate > cfg.TargetRate {
		msg := fmt.Sprintf("🚨 RATE ALERT! 1 USD = %.4f AUD", rate)
		sendTelegram(msg)
		fmt.Println(" Alert sent to Telegram!")
	} else {
		fmt.Println(" Rate is too low. No alert sent.")
	}
}

// api handler for /api/rate
func handleGetRate(w http.ResponseWriter, r *http.Request) {
	// enable cors (so client from a different origin can access)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	mu.RLock() // read lock
	rate := currentRate
	updated := lastUpdated
	mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]any{
		"rate":        rate,
		"lastUpdated": updated.Format(time.RFC3339),
	})
}

// Handler for /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "UP",
		"message": "Conversion Bot is healthy 🚀",
		"time":    time.Now().Format(time.RFC3339),
	})
}
