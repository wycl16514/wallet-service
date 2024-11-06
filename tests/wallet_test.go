package tests

import (
	"bytes"
	"config"
	"encoding/json"
	"fmt"
	"handles"
	"net/http"
	"net/http/httptest"
	"services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var walletService *services.WalletService
var walletHandler *handles.WalletHandler

func setup() {
	// Initialize the wallet service and handler
	db, err := config.InitDB() // Assuming InitDB initializes the PostgreSQL connection
	if err != nil {
		panic(err)
	}

	// tables, err := config.ListTables(db)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(tables)

	walletService = &services.WalletService{DB: db}
	walletHandler = &handles.WalletHandler{Service: walletService}
}

func TestDeposit(t *testing.T) {
	setup()

	// Step 1: Create a user (manually, or assume user creation is handled elsewhere)
	userID := 1 // Assume the user is created with ID 100

	// Step 2: Prepare the request to deposit money
	depositAmount := decimal.NewFromFloat(100.50)

	body := map[string]string{
		"amount": depositAmount.String(),
	}
	bodyJSON, _ := json.Marshal(body)
	userStr := fmt.Sprintf("%d", userID)
	req, err := http.NewRequest(http.MethodPost, "/wallet/"+userStr+"/deposit", bytes.NewReader(bodyJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the result
	rr := httptest.NewRecorder()

	// Create a new Gin engine
	router := gin.Default()
	router.POST("/wallet/:user_id/deposit", walletHandler.Deposit)

	// Step 3: Perform the request
	router.ServeHTTP(rr, req)

	// Step 4: Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Deposit successful")

	// Step 5: Check the updated balance
	var balance decimal.Decimal
	err = walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	isEqual := depositAmount.Equal(balance)
	// Check if the deposit amount is correct
	assert.Equal(t, true, isEqual)
}
