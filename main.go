package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type receipt struct {
	ID           uuid.UUID
	Retailer     string      `json:"retailer"`
	PurchaseDate string      `json:"purchaseDate"`
	PurchaseTime string      `json:"purchaseTime"`
	Total        json.Number `json:"total"`
	Items        []items     `json:"items"`
}

type items struct {
	ShortDescription string      `json:"shortDescription"`
	Price            json.Number `json:"price"`
}

type receiptID struct {
	ID uuid.UUID `json:"id"`
}

type points struct {
	Points int `json:"points"`
}

var receipts = []receipt{}

func main() {
	router := gin.Default()
	router.POST("/receipts/process", processReceipts)
	router.GET("/receipts/:id/points", getPoints)

	router.Run("localhost:8080")
}

func processReceipts(c *gin.Context) {
	var newReceipt receipt

	if err := c.BindJSON(&newReceipt); err != nil {
		fmt.Println(err.Error())
		return
	}

	newReceipt.ID = uuid.New()
	receipts = append(receipts, newReceipt)

	c.IndentedJSON(http.StatusOK, receiptID{ID: newReceipt.ID}) //Example response { "id": "7fb1377b-b223-49d9-a31a-5a02701dd310" }
}

func getPoints(c *gin.Context) {
	var pointTotal int
	var receipt receipt
	id := c.Param("id")

	for _, a := range receipts {
		if a.ID.String() == id {
			receipt = a
		}
	}

	//One point for every alphanumeric character in the retailer name
	for _, a := range receipt.Retailer {
		if unicode.IsLetter(a) || unicode.IsNumber(a) {
			pointTotal += 1
		}
	}
	//50 Points if the total is a round dollar amount with no cents
	totaldollars, err := receipt.Total.Float64()
	cents := int(totaldollars * 100)

	if cents%100 == 0 {
		pointTotal += 50
	}
	//25 points if the total is a multiple of 0.25
	if cents%25 == 0 {
		pointTotal += 25
	}
	//5 points for every two items on the receipt
	pointTotal += ((len(receipt.Items) / 2) * 5)

	//If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer.
	//The result is the number of points earned.
	for _, a := range receipt.Items {
		if len(strings.TrimSpace(a.ShortDescription))%3 == 0 {
			priceFloat, err := a.Price.Float64()
			if err != nil {
				fmt.Println(err)
			}
			pointTotal += int(math.Ceil(priceFloat * 0.2))
		}
	}

	//6 points if the day in the purchase date is odd.
	date, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		fmt.Println(err)
	}
	if date.Day()%2 == 1 {
		pointTotal += 6
	}

	//10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		fmt.Println(err)
	}
	//This is assuming that 2pm is inclusive and 4pm is exclusive
	if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
		pointTotal += 10
	}
	c.IndentedJSON(http.StatusOK, points{Points: pointTotal})
}
