package main

import (
	"config"
	"fmt"
	"services"
)

func main() {
	db, err := config.InitDB()
	if err != nil {
		fmt.Printf("db config err: %v\n", err)
	}
	fmt.Printf("db config ok...")

	service := services.WalletService{DB: db}

	fmt.Printf("init service ok...:%v", service)
}
