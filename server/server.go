package main

import (
	"fmt"

	"github.com/zahra-pzk/Chatbot_Project3/server/admin"
	"github.com/zahra-pzk/Chatbot_Project3/server/user"
)

func main() {
	fmt.Println("Select Client Mode:")
	fmt.Println("1. User Client")
	fmt.Println("2. Admin Client")
	var choice int
	fmt.Print("Enter choice: ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		user.Start()
	case 2:
		admin.Start()
	default:
		fmt.Println("Invalid choice.")
	}
}