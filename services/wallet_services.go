package services

import (
	"database/sql"
	"errors"

	"github.com/shopspring/decimal"
)

type WalletService struct {
	DB *sql.DB
}

/*
In the following code, we use "FOR UPDATE" to lock resources at db level, and
avoid critical resources deadlock
*/

func (s *WalletService) CreateWallet() (int64, error) {
	var id int64
	err := s.DB.QueryRow("INSERT INTO wallets (balance) VALUES ($1) RETURNING id", 0).Scan(&id)
	return id, err
}

func (s *WalletService) Deposit(userID int, amountStr string) error {
	/*
		handle deposit request, parse the number string into decimal value,
	*/
	amount, err := ParseAmount(amountStr)
	if err != nil {
		return err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//get the current balance of given user
	var balance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		return err
	}

	//should we check the amount is poisitive numbe before add to balance?
	newBalance := balance.Add(amount)
	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE user_id = $2", newBalance, userID)
	if err != nil {
		return err
	}

	//save current deposite as transaction record
	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES ($1, 'deposit', $2)", userID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *WalletService) Withdraw(userID int, amountStr string) error {
	amount, err := ParseAmount(amountStr)
	if err != nil {
		return err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//get the current balance for given account
	var balance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		return err
	}

	//make sure withdraw can't more than the amount of balance
	if balance.LessThan(amount) {
		return errors.New("Insufficient balance")
	}

	//reduce the amount from balance and set new balance
	newBalance := balance.Sub(amount)
	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE user_id = $2", newBalance, userID)
	if err != nil {
		return err
	}

	//record current withdraw as a transaction record
	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount) VALUES ($1, 'withdraw', $2)", userID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *WalletService) Transfer(fromUserID, toUserID int, amountStr string) error {
	amount, err := ParseAmount(amountStr)
	if err != nil {
		return err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//get balance for user who want to transfer money
	var fromBalance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE", fromUserID).Scan(&fromBalance)
	if err != nil {
		return err
	}

	//check given user has enough money to transfer
	//should we check the amount is positive number?
	if fromBalance.LessThan(amount) {
		return errors.New("Insufficient balance")
	}

	//get the balance of the receiver
	var toBalance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE", toUserID).Scan(&toBalance)
	if err != nil {
		return err
	}

	//reduce the transfer amount from sender and add to receiver
	newFromBalance := fromBalance.Sub(amount)
	newToBalance := toBalance.Add(amount)

	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE user_id = $2", newFromBalance, fromUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE user_id = $2", newToBalance, toUserID)
	if err != nil {
		return err
	}

	//record this transfer as a transaction record
	_, err = tx.Exec("INSERT INTO transactions (user_id, type, amount, to_user_id) VALUES ($1, 'transfer', $2, $3)", fromUserID, amount, toUserID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *WalletService) GetBalance(userID int) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := s.DB.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

// GetTransactionHistory returns the transaction history for a given user
func (s *WalletService) GetTransactionHistory(userID int) ([]Transaction, error) {
	rows, err := s.DB.Query("SELECT id, user_id, type, amount, to_user_id, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		var amount decimal.Decimal
		err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Type, &amount, &transaction.ToUserID, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transaction.Amount = amount
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
