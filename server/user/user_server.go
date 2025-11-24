package user

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

const (
	apiURL = "http://localhost:8080"
	wsURL  = "ws://localhost:8080"
)

type WSMessage struct {
	Content          string `json:"content"`
	SenderExternalID string `json:"sender_external_id"`
}

func Start() {
	fmt.Println("--- User Client Started ---")
	menu()
}

func menu() {
	for {
		fmt.Println("\n1. Login or Register\n2. Exit")
		var c int
		fmt.Print("Enter choice: ")
		fmt.Scanln(&c)

		switch c {
		case 1:
			loginOrRegister()
		case 2:
			return
		}
	}
}

func loginOrRegister() {
	for {
		fmt.Println("\nPlease choose an option: ")
		fmt.Println("1. Register")
		fmt.Println("2. Login")
		fmt.Println("3. Back")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			register()
		case 2:
			login()
		case 3:
			return
		default:
			fmt.Println("Invalid choice, try again.")
		}
	}
}

func register() {
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("registration failed: %s | Body: %s\n", resp.Status, string(bodyBytes))
		return
	}
	fmt.Println("registration successful")
}

func login() {
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

	var loginResp struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ExternalID string `json:"user_external_id"`
			Name       string `json:"name"`
			Role       string `json:"role"`
		} `json:"user"`
	}

	respBodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		fmt.Println("Error decoding login response:", err)
		return
	}

	token := loginResp.AccessToken
	extID := loginResp.User.ExternalID

	if extID == "" {
		fmt.Println("CRITICAL ERROR: Could not find user ExternalID in response!")
		return
	}
	fmt.Println("Login success. Welcome,", loginResp.User.Name)

	runUserPanel(token, extID)
}

func runUserPanel(token, userID string) {
	for {
		fmt.Println("\n1. Connect to Support\n2. Logout")
		var c int
		fmt.Print("enter choice: ")
		fmt.Scanln(&c)

		switch c {
		case 1:
			connectSupport(token, userID)
		case 2:
			return
		}
	}
}

func connectSupport(token, userID string) {
	fmt.Println("Creating support chat...")

	reqBody := []byte(`{"content": "New support request started", "name": "Support Request"}`)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL+"/chats", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to create chat:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to create chat, status: %s | Error: %s\n", resp.Status, string(bodyBytes))
		return
	}

	var chatResp struct {
		ExternalID string `json:"chat_external_id"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		fmt.Println("Error decoding chat response:", err)
		return
	}

	chatID := chatResp.ExternalID
	fmt.Printf("Chat created (ID: %s). Connecting...\n", chatID)

	u, err := url.Parse(wsURL + "/ws/chats/" + chatID)
	if err != nil {
		log.Println("URL parse error:", err)
		return
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		fmt.Printf("WebSocket connection failed: %v\n", err)
		return
	}
	defer wsConn.Close()

	fmt.Println("Connected! Type messages below ('exit' to quit):")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				return
			}

			var msgObj WSMessage
			if err := json.Unmarshal(message, &msgObj); err == nil {
				if msgObj.SenderExternalID != userID {
					fmt.Printf("\n[Support]: %s\n> ", msgObj.Content)
				}
			} else {
				fmt.Printf("\n[System]: %s\n> ", message)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" {
			wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		if text == "" {
			fmt.Print("> ")
			continue
		}
		err := wsConn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			return
		}
		fmt.Print("> ")
	}
	<-done
}