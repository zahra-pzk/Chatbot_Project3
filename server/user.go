package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zahra-pzk/Chatbot_Project3/server/admin"
)

const apiURL = "http://localhost:8080"

func menu(apiURL string) {
	for {
		fmt.Println("\n1. login or register\n2. connect to support")
		var c int
		fmt.Print("enter choice: ")
		fmt.Scanln(&c)

		switch c {
		case 1:
			loginOrRegister(apiURL)
		case 2:
			fmt.Println("support not ready yet")
		}
	}
}

func loginOrRegister(apiURL string) {
	for {
		fmt.Println("\nplease choose an option: ")
		fmt.Println("1. register")
		fmt.Println("2. login")
		fmt.Println("3. exit")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			register(apiURL)
		case 2:
			login(apiURL)
		case 3:
			return
		default:
			fmt.Println("Invalid choice, try again.")
		}
	}
}

func register(apiURL string) {
	var name, username, email, phone, password, confirm string

	fmt.Print("Enter your full name: ")
	fmt.Scanln(&name)

	fmt.Print("Enter a username: ")
	fmt.Scanln(&username)

	fmt.Print("Enter your phone number: ")
	fmt.Scanln(&phone)

	fmt.Print("Enter your email: ")
	fmt.Scanln(&email)

	fmt.Print("Enter password: ")
	fmt.Scanln(&password)

	fmt.Print("Confirm password: ")
	fmt.Scanln(&confirm)

	if password != confirm {
		fmt.Println("passwords do not match")
		return
	}

	body := map[string]string{
		"name":         name,
		"username":     username,
		"phone_number": phone,
		"email":        email,
		"password":     password,
		"role":         "user",
	}

	b, _ := json.Marshal(body)

	resp, err := http.Post(apiURL+"/users", "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("registration failed:", resp.Status)
		return
	}

	fmt.Println("registration successful")
}

func login(apiURL string) {
	var username, password string

	fmt.Print("Enter username: ")
	fmt.Scanln(&username)

	fmt.Print("Enter password: ")
	fmt.Scanln(&password)

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(apiURL+"/users/login", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("error connecting:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("login failed:", resp.Status)
		return
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
		ExternalID  string `json:"external_id"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	token := loginResp.AccessToken
	extID := loginResp.ExternalID

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL+"/users/info/"+extID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	userResp, err := client.Do(req)
	if err != nil {
		fmt.Println("failed to fetch user info:", err)
		return
	}
	defer userResp.Body.Close()

	var userData struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		Username string `json:"username"`
	}

	json.NewDecoder(userResp.Body).Decode(&userData)

	fmt.Println("Welcome,", userData.Name)

	switch userData.Role {
	case "admin", "superadmin", "system":
		runAdminServer()
	default:
		runUserServer()
	}
}

func runAdminServer() {
	admin.Start()
}

func runUserServer() {
	for {
		fmt.Println("\n1. connect to support\n2. logout")
		var c int
		fmt.Print("enter choice: ")
		fmt.Scanln(&c)

		switch c {
		case 1:
			
		case 2:
			fmt.Println("cannot logout in this moment")
		}
	}
}
func main() {
	fmt.Println("user client started…")
	menu(apiURL)
}
