package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getDataBaseConnection() (*sql.DB, error) {
	dsn := "root:Vishnu@tj@tcp(localhost:3306)/inventory"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err

	}
	return db, db.Ping()

}
func getUserRole(db *sql.DB, username, password string) (string, error) {
	query := "SELECT role FROM users WHERE username = ? AND password = ?"
	row := db.QueryRow(query, username, password)

	var role string
	err := row.Scan(&role)
	if err != nil {
		return "", nil
	}
	return role, nil
}
func getProductDetails(db *sql.DB, itemID string) *sql.Row {
	query := "SELECT * FROM productdetails WHERE item_id = ?"
	return db.QueryRow(query, itemID)
}
func getGstRate(db *sql.DB, itemType string) (int, error) {
	query := "SELECT GST From GST WHERE type =?"
	row := db.QueryRow(query, itemType)
	var gst int
	if err := row.Scan(&gst); err != nil {
		return 5, nil
	}
	return gst, nil

}
func getOffer(db *sql.DB, itemID string) (string, error) {
	query := "SELECT offer FROM OFFER WHERE item_id = ?"
	row := db.QueryRow(query, itemID)

	var offer string
	err := row.Scan(&offer)
	if err != nil {
		return "no offer available", nil
	}
	return offer, nil
}
func calculateFinalPrice(itemPrice int, offer string, gstRate int) (float64, float64, float64) {
	var discount float64
	offer = strings.TrimSpace(strings.ToLower(offer))

	switch offer {
	case "10% off":
		discount = 0.10
	case "5% off":
		discount = 0.05
	case "3% off":
		discount = 0.03
	default:
		discount = 0
	}
	discountedPrice := float64(itemPrice) * (1 - discount)
	gstAmount := discountedPrice * float64(gstRate) / 100
	finalAmount := discountedPrice + gstAmount

	return discountedPrice, gstAmount, finalAmount
}
func updateStock(db *sql.DB, itemID string, quantitySold int) error {
	query := "UPDATE productdetails SET item_quantity = item_quantity - ? WHERE item_id = ?"
	_, err := db.Exec(query, quantitySold, itemID)
	return err
}
func refillStock(db *sql.DB, itemID string, quantityToAdd int) error {
	query := "UPDATE productdetails SET item_quantity = item_quantity + ? WHERE item_id = ?"
	_, err := db.Exec(query, quantityToAdd, itemID)
	return err
}
func removeExpiredItem(db *sql.DB, itemID string) error {
	query := "DELETE FROM productdetails WHERE item_id = ?"
	_, err := db.Exec(query, itemID)
	return err
}
func getBillingDetails(db *sql.DB, itemID string, quantitySold int) (float64, error) {
	row := getProductDetails(db, itemID)

	var itemName, itemType string
	var itemPrice, itemQuantity int
	var expiryDate int64
	var dbItemID string

	err := row.Scan(&dbItemID, &itemName, &itemType, &itemPrice, &itemQuantity, &expiryDate)
	if err != nil {
		fmt.Println("Product not found with item ID:", itemID)
		return 0.0, nil
	}

	currentEpoch := time.Now().Unix()
	if expiryDate < currentEpoch {
		fmt.Printf("Item %s is expired and removed.\n", itemID)
		_ = removeExpiredItem(db, itemID)
		return 0.0, nil
	}

	if itemQuantity < quantitySold {
		fmt.Printf("Insufficient stock. Available: %d\n", itemQuantity)
		return 0.0, nil
	}

	gstRate, _ := getGstRate(db, itemType)
	offer, _ := getOffer(db, itemID)
	discounted, gstAmount, finalAmount := calculateFinalPrice(itemPrice, offer, gstRate)

	fmt.Printf("\nItem ID       : %s\n", itemID)
	fmt.Printf("Product       : %s\n", itemName)
	fmt.Printf("Category      : %s\n", itemType)
	fmt.Printf("Price         : ₹%d\n", itemPrice)
	fmt.Printf("GST           : %d%%\n", gstRate)
	fmt.Printf("Offer         : %s\n", offer)
	fmt.Printf("Discounted    : ₹%.2f\n", discounted)
	fmt.Printf("GST Amount    : ₹%.2f\n", gstAmount)
	fmt.Printf("Quantity Sold : %d\n", quantitySold)
	fmt.Printf("Final Amount  : ₹%.2f\n\n", finalAmount)

	_ = updateStock(db, itemID, quantitySold)
	return finalAmount, nil
}
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var db *sql.DB
	var err error

	for {
		fmt.Print("Enter username: ")
		scanner.Scan()
		username := scanner.Text()

		fmt.Print("Enter password: ")
		scanner.Scan()
		password := scanner.Text()

		db, err = getDataBaseConnection()
		if err != nil {
			log.Fatal("DB connection failed:", err)
		}

		role, err := getUserRole(db, username, password)
		if err != nil || role == "" {
			fmt.Println("Invalid username or password. Please try again.")
			continue
		}

		fmt.Println("Login successful as", role)

		if strings.ToLower(role) == "user" {
			totalBill := 0.0
			for {
				fmt.Print("Enter item ID: ")
				scanner.Scan()
				itemID := scanner.Text()

				fmt.Print("Enter quantity sold: ")
				scanner.Scan()
				qty, _ := strconv.Atoi(scanner.Text())

				amount, _ := getBillingDetails(db, itemID, qty)
				totalBill += amount

				fmt.Print("Do you want to bill another item? (yes/no): ")
				scanner.Scan()
				if strings.ToLower(scanner.Text()) != "yes" {
					break
				}
			}
			fmt.Printf("Total Bill Amount: ₹%.2f\n", totalBill)

		} else if strings.ToLower(role) == "admin" {
			for {
				fmt.Print("Enter item ID to refill: ")
				scanner.Scan()
				itemID := scanner.Text()

				fmt.Print("Enter quantity to add: ")
				scanner.Scan()
				qty, _ := strconv.Atoi(scanner.Text())

				err := refillStock(db, itemID, qty)
				if err != nil {
					fmt.Println("Refill failed:", err)
				} else {
					fmt.Println("Stock refilled successfully.")
				}

				fmt.Print("Refill another item? (yes/no): ")
				scanner.Scan()
				if strings.ToLower(scanner.Text()) != "yes" {
					break
				}
			}
		}

		break
	}

	db.Close()
}
