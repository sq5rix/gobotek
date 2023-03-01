package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"github.com/suquant/kucoin-go-sdk/kucoin"
	"github.com/markcheno/go-talib"
)

const (
	tradingSymbol = "ETH-USDT"
	timeFrame     = "15m"
)

var (
	client      *kucoin.Client
	usdtBalance float64
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize KuCoin client
	client = kucoin.NewClient(os.Getenv("KUCOIN_API_KEY"), os.Getenv("KUCOIN_API_SECRET"), os.Getenv("KUCOIN_PASSPHRASE"))

	// Main loop
	for {
		buy()
		sell()
		time.Sleep(1 * time.Minute)
	}
}

func checkMarketCondition() bool {
	// Get market data
	klines, err := client.Klines(tradingSymbol, timeFrame, 100)
	if err != nil {
		log.Fatalf("Failed to get klines: %v", err)
	}

	// Calculate MACD and SMA
	macd, signal, _ := talib.Macd(klines.Close, 12, 26, 9)
	sma, _ := talib.Sma(klines.Close, 20)

	// Check condition
	return macd[len(macd)-1] < signal[len(signal)-1] && sma[len(sma)-1] > sma[len(sma)-2]
}

func checkBalance() {
	// Get USDT balance
	accounts, err := client.AccountList()
	if err != nil {
		log.Fatalf("Failed to get account list: %v", err)
	}
	for _, account := range accounts {
		if account.Currency == "USDT" && account.Type == "trade" {
			usdtBalance, err = strconv.ParseFloat(account.Available, 64)
			if err != nil {
				log.Fatalf("Failed to parse USDT balance: %v", err)
			}
			break
		}
	}
}

func buy() {
	// Check market condition
	if !checkMarketCondition() {
		return
	}

	// Check balance
	checkBalance()
	if usdtBalance <= 0 {
		return
	}

	// Calculate order size based on available USDT balance
	price, err := client.GetLastPrice(tradingSymbol)
	if err != nil {
		log.Fatalf("Failed to get last price: %v", err)
	}
	orderSize := usdtBalance / price

	// Place buy order
	order, err := client.CreateOrder(tradingSymbol, kucoin.BUY, "", fmt.Sprintf("%.8f", orderSize), fmt.Sprintf("%.8f", price))
	if err != nil {
		log.Fatalf("Failed to place buy order: %v", err)
	}
	log.Printf("Placed buy order: %v", order)

	// Update USDT balance
	usdtBalance -= orderSize * price
}

func sell(client *KuCoinClient, ticker string) error {
	balance, err := GetBalance(client, ticker)
	if err != nil {
		return err
	}

	if balance.Free == "0" {
		log.Println("No ETH to sell")
		return nil
	}

	price, err := GetTickerPrice(client, ticker)
	if err != nil {
		return err
	}

	orderPrice := price * 0.99 // Sell at 1% below market price

	orderSize, err := strconv.ParseFloat(balance.Free, 64)
	if err != nil {
		return err
	}

	log.Printf("Selling %v ETH at %v USDT/ETH...\n", orderSize, orderPrice)

	orderId, err := client.PlaceLimitSellOrder(ticker, orderSize, orderPrice)
	if err != nil {
		return err
	}

	log.Printf("Order placed: %v\n", orderId)

	return nil
}

