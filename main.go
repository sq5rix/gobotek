package main

import (
	"log"
	"strconv"
	"strings"
	"time"
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

func Buy(client *KuCoinClient, tradingSymbol string) error {
	ticker := strings.Split(tradingSymbol, "/")[0]
	balance, err := GetBalance(client, "USDT")
	if err != nil {
		return err
	}

	if balance.Free == "0" {
		log.Println("No USDT to buy")
		return nil
	}

	price, err := GetTickerPrice(client, tradingSymbol)
	if err != nil {
		return err
	}

	orderPrice := price * 1.01 // Buy at 1% above market price

	orderSize := (FloatFromString(balance.Free) / FloatFromString(price)) * 0.99 // Buy 99% of available balance

	log.Printf("Buying %v %s at %v USDT/%s...\n", orderSize, ticker, orderPrice, ticker)

	orderId, err := client.PlaceLimitBuyOrder(tradingSymbol, orderSize, orderPrice)
	if err != nil {
		return err
	}

	log.Printf("Order placed: %v\n", orderId)

	return nil
}

func Sell(client *KuCoinClient, tradingSymbol string) error {
	ticker := strings.Split(tradingSymbol, "/")[0]
	balance, err := GetBalance(client, ticker)
	if err != nil {
		return err
	}

	if balance.Free == "0" {
		log.Printf("No %s to sell", ticker)
		return nil
	}

	price, err := GetTickerPrice(client, tradingSymbol)
	if err != nil {
		return err
	}

	orderPrice := price * 0.99 // Sell at 1% below market price

	orderSize, err := strconv.ParseFloat(balance.Free, 64)
	if err != nil {
		return err
	}

	log.Printf("Selling %v %s at %v USDT/%s...\n", orderSize, ticker, orderPrice, ticker)

	orderId, err := client.PlaceLimitSellOrder(tradingSymbol, orderSize, orderPrice)
	if err != nil {
		return err
	}

	log.Printf("Order placed: %v\n", orderId)

	return nil
}

func GetBalance(client *KuCoinClient, currency string) (*Balance, error) {
	account, err := client.GetAccount()
	if err != nil {
		return nil, err
	}

	for _, balance := range account.Balances {
		if balance.Currency == currency {
			return &balance, nil
		}
	}

	return nil, fmt.Errorf("balance for %s not found", currency)
}

func checkMarketCondition(client *KuCoinClient, tradingSymbol string) (bool, error) {
	ohlcv, err := client.GetKlines(tradingSymbol, "15min", 50)
	if err != nil {
		return false, err
	}

	// Calculate the MACD indicator
	macd, err := talib.Macd(ohlcv.Close, 12, 26, 9)
	if err != nil {
		return false, err
	}

	// Calculate the 10-period simple moving average
	sma10, err := talib.Sma(ohlcv.Close, 10)
	if err != nil {
		return false, err
	}

	// Check if MACD is below zero and the SMA10 is above the last candle's close
	return macd.MACD[len(macd.MACD)-1] < 0 && sma10[len(sma10)-1] > ohlcv.Close[len(ohlcv.Close)-1], nil
}

