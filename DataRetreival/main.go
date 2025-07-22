package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ExchangeResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter base currency (e.g., INR): ")
	baseCurrency, _ := reader.ReadString('\n')
	baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))

	apiURL := fmt.Sprintf("https://api.frankfurter.dev/v1/latest?base=%s", baseCurrency)

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error fetching exchange rates:", err)
		return
	}
	defer resp.Body.Close()

	var data ExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	file, err := os.Create("stronger_currencies.csv")
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	fmt.Println("\nStronger currencies than", baseCurrency)
	fmt.Println("Currency\tRate")
	writer.Write([]string{"Currency", "Rate"})

	for currency, rate := range data.Rates {
		if rate < 1 {
			fmt.Printf("%s\t\t%.4f\n", currency, rate)
			writer.Write([]string{currency, fmt.Sprintf("%.4f", rate)})
		}
	}

	fmt.Println("\n Data written to stronger_currencies.csv")
}
