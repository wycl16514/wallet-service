package models

import "github.com/shopspring/decimal"

type Wallet struct {
	ID      int             `json:"id"`
	UserID  int             `json:"user_id"`
	Balance decimal.Decimal `json:"balance"`
}

type Transaction struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Type      string          `json:"type"`
	Amount    decimal.Decimal `json:"amount"`
	ToUserID  int             `json:"to_user_id,omitempty"`
	CreatedAt string          `json:"created_at"`
}

func ParseAmount(amountStr string) (decimal.Decimal, error) {
	/*
		convert decimal string to decimal value
	*/
	return decimal.NewFromString(amountStr)
}
