package admin

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	apiURL = "http://localhost:8080"
	wsURL  = "ws://localhost:8080"
)

type WSMessage struct {
	Content string `json:"content"`
}

type ChatItem struct {
	ChatExternalID string `json:"chat_external_id"`
	UserExternalID string `json:"user_external_id"`
	Status         string `json:"status"`
	UpdatedAt      string `json:"updated_at"`
}

var currentChats []ChatItem
var mu sync.Mutex

func Start() {
	fmt.Println("--- Admin Client Started ---")
	menu()
}

func menu() {
	for {
		fmt.Println("\n1. Login (Admin)\n2. Register (Admin)\n3. Exit")
		var c int
		fmt.Print("Enter choice: ")
		fmt.Scanln(&c)

		switch c {
		case 1:
			loginAdmin()
		case 2:
			registerAdmin()
		case 3:
			return
		default:
			fmt.Println("Invalid choice. Try again.")
		}
	}
}

func loginAdmin() {
	var username, password string
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password: ")
	fmt.Scanln(&password)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := http.Post(apiURL+"/users/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Login failed (Status %d): %s\n", resp.StatusCode, string(body))
		return
	}

	var result struct {
		AccessToken string `json:"access_token"`
		User        struct {
			Role string `json:"role"`
		} `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.User.Role != "admin" && result.User.Role != "superadmin" {
		fmt.Println("Login successful, but you are not authorized as an admin/superadmin.")
		return
	}

	fmt.Println("Login successful! Welcome, Admin.")
	selectChat(result.AccessToken)
}

func registerAdmin() {
	var name, username, password, email string
	fmt.Print("Name: ")
	fmt.Scanln(&name)
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password (min 8 chars): ")
	fmt.Scanln(&password)
	fmt.Print("Email (optional): ")
	fmt.Scanln(&email)

	reqBody, _ := json.Marshal(map[string]string{
		"name":     name,
		"username": username,
		"password": password,
		"email":    email,
		"role":     "admin",
	})

	resp, err := http.Post(apiURL+"/users", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Registration failed (Status %d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Println("Admin registration successful! Please log in.")
}

func getChats(token string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL+"/admin/chats", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch chats (Status %d): %s", resp.StatusCode, string(body))
	}

	var chats []ChatItem
	if err := json.NewDecoder(resp.Body).Decode(&chats); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	sort.Slice(chats, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, chats[i].UpdatedAt)
		t2, _ := time.Parse(time.RFC3339, chats[j].UpdatedAt)
		return t1.After(t2)
	})

	mu.Lock()
	currentChats = chats
	mu.Unlock()

	return nil
}

func selectChat(token string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		if err := getChats(token); err != nil {
			fmt.Printf("\nError fetching chats: %v\n", err)
			return
		}

		mu.Lock()
		chats := currentChats
		mu.Unlock()

		fmt.Println("\n--- Available Chats (Refresh: 5) ---")
		if len(chats) == 0 {
			fmt.Println("No active chats found.")
		} else {
			for i, chat := range chats {
				fmt.Printf("[%d] | Status: %s | User ID: %s | Last Update: %s | Chat ID: %s\n",
					i+1, chat.Status, chat.UserExternalID[:8]+"...", chat.UpdatedAt, chat.ChatExternalID[:8]+"...")
			}
		}
		fmt.Println("[R] Refresh | [B] Back to Menu | [X] Exit")
		fmt.Print("Enter chat number to join or option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "b" {
			return
		}
		if strings.ToLower(input) == "r" {
			continue
		}
		if strings.ToLower(input) == "x" {
			os.Exit(0)
		}

		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(chats) {
			fmt.Println("Invalid selection or option.")
			continue
		}

		selectedChat := chats[selection-1]
		chatWithUser(token, selectedChat.ChatExternalID)
	}
}

func chatWithUser(token string, chatID string) {
	fmt.Printf("\nConnecting to chat %s...\n", chatID)

	u, _ := url.Parse(wsURL + "/ws/chats/" + chatID)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		fmt.Printf("Chat WS connection failed: %v\n", err)
		return
	}
	defer wsConn.Close()

	fmt.Println("Connected! (Type 'exit' to leave, or 'close' to close chat)")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("\n[Disconnected] Error in reading msg: %v\n", err)
				}
				return
			}

			var msgObj WSMessage
			if err := json.Unmarshal(message, &msgObj); err == nil {
				fmt.Printf("\n[User]: %s\n> ", msgObj.Content)
			} else {
				fmt.Printf("\n[Raw Message]: %s\n> ", message)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()

		if strings.ToLower(text) == "exit" {
			break
		}
		if strings.ToLower(text) == "close" {
			fmt.Println("Chat closed locally. Please remember to close it via API if needed.")
			break
		}

		msgBody := map[string]string{"content": text}
		jsonMsg, _ := json.Marshal(msgBody)

		err = wsConn.WriteMessage(websocket.TextMessage, jsonMsg)
		if err != nil {
			fmt.Printf("Write failed: %v\n", err)
			break
		}
		fmt.Print("> ")
	}

	wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))

	<-done
	fmt.Println("\nChat session ended. Returning to chat list.")
}
