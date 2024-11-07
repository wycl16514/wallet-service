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

	//need to make sure user with id 1 exists
	userID := 1

	var currentDeposit decimal.Decimal
	err := walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&currentDeposit)
	if err != nil {
		t.Fatal(err)
	}
	// Prepare the request to deposit money
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

	totalDeposit := currentDeposit.Add(depositAmount)

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

	isEqual := totalDeposit.Equal(balance)
	// Check if the deposit amount is correct
	assert.Equal(t, true, isEqual)
}

func TestWithdraw(t *testing.T) {
	setup()

	// make sure user with id 1 exists
	//and his balance is more than 50.00
	userID := 1

	//get his initial deposit
	var initialDeposit decimal.Decimal
	err := walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&initialDeposit)
	if err != nil {
		t.Fatal(err)
	}

	//Prepare the request to withdraw money
	withdrawAmount := decimal.NewFromFloat(50.00)
	body := map[string]string{
		"amount": withdrawAmount.String(),
	}
	bodyJSON, _ := json.Marshal(body)
	userStr := fmt.Sprintf("%d", userID)
	req, err := http.NewRequest(http.MethodPost, "/wallet/"+userStr+"/withdraw", bytes.NewReader(bodyJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the result
	rr := httptest.NewRecorder()

	// Create a new Gin engine
	router := gin.Default()
	router.POST("/wallet/:user_id/withdraw", walletHandler.Withdraw)

	// Step 3: Perform the request
	router.ServeHTTP(rr, req)

	// Step 4: Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Withdrawal successful")

	// Step 5: Check the updated balance
	var balance decimal.Decimal
	err = walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatal(err)
	}

	// The balance should be the initial deposit minus the withdrawal
	expectedBalance := initialDeposit.Sub(withdrawAmount)
	isEqual := expectedBalance.Equal(balance)
	assert.Equal(t, true, isEqual)
}

func TestGetBalance(t *testing.T) {
	setup()

	// make sure user with id 1 exists
	userID := 1
	//init user with given balance
	initialDeposit := decimal.NewFromFloat(300.00)
	_, err := walletService.DB.Exec("UPDATE wallets set balance = $2 where user_id = $1", userID, initialDeposit)
	if err != nil {
		t.Fatal(err)
	}

	//  Prepare the request to get the balance
	userStr := fmt.Sprintf("%d", userID)
	req, err := http.NewRequest(http.MethodGet, "/wallet/"+userStr+"/balance", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the result
	rr := httptest.NewRecorder()

	// Create a new Gin engine
	router := gin.Default()
	router.GET("/wallet/:user_id/balance", walletHandler.GetBalance)

	//Perform the request
	router.ServeHTTP(rr, req)

	//Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), initialDeposit.String())
}

func TestTransfer(t *testing.T) {
	setup()

	// make sure user with id 1 and 2 exist
	fromUserID := 1
	toUserID := 2

	//get initial deposit of sender for later verification
	var initialDeposit decimal.Decimal
	err := walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", fromUserID).Scan(&initialDeposit)
	if err != nil {
		t.Fatal(err)
	}
	//set receiver balance to 0 for later verification
	_, err = walletService.DB.Exec("UPDATE wallets set balance = $2 where user_id = $1", toUserID, decimal.Zero)
	if err != nil {
		t.Fatal(err)
	}

	// Prepare the request to transfer money
	transferAmount := decimal.NewFromFloat(50.00)
	body := map[string]interface{}{
		"to_user_id": toUserID,
		"amount":     transferAmount.String(),
	}
	bodyJSON, _ := json.Marshal(body)
	fromUserIDStr := fmt.Sprintf("%d", fromUserID)
	req, err := http.NewRequest(http.MethodPost, "/wallet/"+fromUserIDStr+"/transfer", bytes.NewReader(bodyJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the result
	rr := httptest.NewRecorder()

	// Create a new Gin engine
	router := gin.Default()
	router.POST("/wallet/:from_user_id/transfer", walletHandler.Transfer)

	//  Perform the request
	router.ServeHTTP(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Transfer successful")

	// Check the updated balances
	var fromUserBalance decimal.Decimal
	var toUserBalance decimal.Decimal
	err = walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", fromUserID).Scan(&fromUserBalance)
	if err != nil {
		t.Fatal(err)
	}
	err = walletService.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", toUserID).Scan(&toUserBalance)
	if err != nil {
		t.Fatal(err)
	}

	// The sender's balance should be decreased by transferAmount
	expectedFromUserBalance := initialDeposit.Sub(transferAmount)
	isEqual := fromUserBalance.Equal(expectedFromUserBalance)
	assert.Equal(t, true, isEqual)

	// The receiver's balance should be increased by transferAmount
	isEqual = transferAmount.Equal(toUserBalance)
	assert.Equal(t, true, isEqual)
}
