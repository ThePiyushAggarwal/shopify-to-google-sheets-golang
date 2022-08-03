package main

import (
	"os"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"io"
	"io/ioutil"
	"gopkg.in/Iwark/spreadsheet.v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"encoding/json"
	"strconv"
)

func main() {
	// Loading .env File
	err := godotenv.Load(".env")
  if err != nil {
    log.Fatal("Error loading .env file")
  }

	// Assigning Environment variables
	admin_access_token := os.Getenv("ADMIN_ACCESS_TOKEN")
	shopify_url := os.Getenv("SHOPIFY_URL")
	spreadsheet_id := os.Getenv("SPREADSHEET_ID")

	// Requesting data from Shopify API
	client := http.Client{}
	req, err := http.NewRequest("GET", shopify_url+"/admin/api/2022-07/orders.json", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = http.Header{
		"Content-Type":           {"application/json"},
		"X-Shopify-Access-Token": {admin_access_token},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Using google sheets API
	data, err := ioutil.ReadFile("client_secret.json")
	checkError(err)
	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	checkError(err)
	clientx := conf.Client(context.TODO())

	service := spreadsheet.NewServiceWithClient(clientx)
	spreadsheet, err := service.FetchSpreadsheet(spreadsheet_id)
	checkError(err)
	sheet, err := spreadsheet.SheetByIndex(0)
	checkError(err)

	// Setting the fields
	fields_required := []string{"id", "browser_ip", "current_subtotal_price", "email", "gateway", "order_number"}
	
	for i, j := range fields_required {
		sheet.Update(0, i, j)
	}

	var orders Orders
	json.Unmarshal(body, &orders)

	for i, j:= range orders.Orders {
		sheet.Update(i+1, 0, strconv.Itoa(j.Id))
		sheet.Update(i+1, 1, j.BrowserIp)
		sheet.Update(i+1, 2, strconv.Itoa(j.CurrentSubtotalPrice))
		sheet.Update(i+1, 3, (j.Email))
		sheet.Update(i+1, 4, (j.Gateway))
		sheet.Update(i+1, 5, strconv.Itoa(j.OrderNumber))
	}

	err = sheet.Synchronize()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type Orders struct {
	Orders []Order
}

type Order struct {
	Id int
	BrowserIp string `json:"browser_ip"`
	CurrentSubtotalPrice int `json:"current_subtotal_price"`
	Email string 
	Gateway string
	OrderNumber int `json:"order_number"`
}