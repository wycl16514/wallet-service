package main

import (
	"config"
	"fmt"
)

func main() {
	_, err := config.InitDB()
	if err != nil {
		fmt.Printf("db config err: %v\n", err)
	}
	fmt.Printf("db config ok...")
}
