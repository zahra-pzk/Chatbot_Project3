package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bufio"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ledongthuc/pdf"
	"github.com/zahra-pzk/Chatbot_Project3/util"
)

type Chunk struct {
	ID        int
	Text      string
	Embedding []float64
}

type WSMessage struct {
	Content          string `json:"content"`
	SenderExternalID string `json:"sender_external_id"`
}

type ChatItem struct {
	ChatExternalID string `json:"chat_external_id"`
	Status         string `json:"status"`
}

var activeChats sync.Map


func loadTextFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var b strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		b.WriteString(scanner.Text())
		b.WriteString("\n")
	}
	return b.String(), scanner.Err()
}

func loadPDFFile(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var b strings.Builder
	totalPage := r.NumPage()
	for i := 1; i <= totalPage; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		txt, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		b.WriteString(txt)
		b.WriteString("\n")
	}
	return b.String(), nil
}

func splitText(text string, chunkSize, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return chunks
	}
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize / 2
	}
	for start := 0; start < n; start += step {
		end := start + chunkSize
		if end > n {
			end = n
		}
		chunks = append(chunks, string(runes[start:end]))
		if end == n {
			break
		}
	}
	return chunks
}

func getEmbeddings(texts []string) ([][]float64, error) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	type EmbReq struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	type EmbResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	payload := EmbReq{
		Model: "text-embedding-ada-002",
		Input: texts,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config.AIBaseURL+"/embeddings", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AIAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("embeddings API error: status %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	var er EmbResp
	if err := json.Unmarshal(bodyBytes, &er); err != nil {
		return nil, err
	}

	out := make([][]float64, len(texts))
	for _, d := range er.Data {
		out[d.Index] = d.Embedding
	}
	return out, nil
}

func chatCompletion(prompt string, contextChunks []Chunk) (string, error) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	var ctxBuilder strings.Builder
	ctxBuilder.WriteString("Use the following context to answer the user's question. If unsure, say you don't know.\n\n")
	for i, c := range contextChunks {
		ctxBuilder.WriteString(fmt.Sprintf("Context %d:\n%s\n\n", i+1, c.Text))
	}
	ctxBuilder.WriteString("User question:\n")
	ctxBuilder.WriteString(prompt)

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type ChatReq struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature"`
	}
	type ChatResp struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	reqObj := ChatReq{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant. Answer in Persian."},
			{Role: "user", Content: ctxBuilder.String()},
		},
		Temperature: 0.0,
	}
	b, _ := json.Marshal(reqObj)
	req, err := http.NewRequest("POST", config.AIBaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AIAPIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("chat API error: status %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	var cr ChatResp
	if err := json.Unmarshal(bodyBytes, &cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from chat API")
	}
	return cr.Choices[0].Message.Content, nil
}

func ensureChunksTable(ctx context.Context, pool *pgxpool.Pool) error {
	create := `
    CREATE TABLE IF NOT EXISTS chunks (
      id SERIAL PRIMARY KEY,
      text TEXT NOT NULL,
      embedding JSONB NOT NULL
    );
    `
	_, err := pool.Exec(ctx, create)
	return err
}

func saveChunksToPostgres(ctx context.Context, pool *pgxpool.Pool, texts []string, embeddings [][]float64) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, t := range texts {
		embJSON, _ := json.Marshal(embeddings[i])
		_, err := tx.Exec(ctx, "INSERT INTO chunks (text, embedding) VALUES ($1, $2)", t, embJSON)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func loadAllChunksFromPostgres(ctx context.Context, pool *pgxpool.Pool) ([]Chunk, error) {
	rows, err := pool.Query(ctx, "SELECT id, text, embedding FROM chunks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chunk
	for rows.Next() {
		var id int
		var text string
		var embJSON []byte
		if err := rows.Scan(&id, &text, &embJSON); err != nil {
			return nil, err
		}
		var emb []float64
		if err := json.Unmarshal(embJSON, &emb); err != nil {
			return nil, err
		}
		out = append(out, Chunk{ID: id, Text: text, Embedding: emb})
	}
	return out, nil
}

func cosineSim(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return -1
	}
	var dot, na, nb float64
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return -1
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func retrieveFromPostgres(ctx context.Context, pool *pgxpool.Pool, query string, topK int) ([]Chunk, error) {
	embs, err := getEmbeddings([]string{query})
	if err != nil {
		return nil, err
	}
	qemb := embs[0]
	chunks, err := loadAllChunksFromPostgres(ctx, pool)
	if err != nil {
		return nil, err
	}
	type scored struct {
		c Chunk
		s float64
	}
	var list []scored
	for _, c := range chunks {
		s := cosineSim(qemb, c.Embedding)
		list = append(list, scored{c: c, s: s})
	}
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].s > list[i].s {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	limit := topK
	if limit > len(list) {
		limit = len(list)
	}
	out := make([]Chunk, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, list[i].c)
	}
	return out, nil
}

func CreateVectorStore(ctx context.Context, pool *pgxpool.Pool, filePath string) error {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	fmt.Println("Processing data...")
	ext := strings.ToLower(filepath.Ext(filePath))
	var text string
	if ext == ".txt" {
		text, err = loadTextFile(filePath)
	} else if ext == ".pdf" {
		text, err = loadPDFFile(filePath)
	} else {
		return fmt.Errorf("unsupported file format")
	}
	if err != nil {
		return err
	}

	chunks := splitText(text, int(config.ChunkSize), int(config.ChunkOverlap))
	fmt.Printf("Created %d chunks\n", len(chunks))

	batchSize := 16
	var allEmb [][]float64
	for i := 0; i < len(chunks); i += batchSize {
		j := i + batchSize
		if j > len(chunks) {
			j = len(chunks)
		}
		batch := chunks[i:j]
		embs, err := getEmbeddings(batch)
		if err != nil {
			return err
		}
		allEmb = append(allEmb, embs...)
		time.Sleep(200 * time.Millisecond)
	}
	if len(allEmb) != len(chunks) {
		return fmt.Errorf("embeddings count mismatch")
	}

	if err := ensureChunksTable(ctx, pool); err != nil {
		return err
	}
	fmt.Println("Saving chunks to Postgres...")
	if err := saveChunksToPostgres(ctx, pool, chunks, allEmb); err != nil {
		return err
	}
	fmt.Println("Vector Store created successfully.")
	return nil
}

func registerBot() {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	payload := map[string]string{
		"name":         "chatbot",
		"username":     config.BotUsername,
		"password":     config.BotPass,
		"email":        "bot@example.com",
		"phone_number": "0000000000",
		"role":         "admin",
	}
	b, _ := json.Marshal(payload)
	http.Post(config.APIURL+"/users", "application/json", bytes.NewBuffer(b))
}

func loginBot() (string, string) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	payload := map[string]string{
		"username": config.BotUsername,
		"password": config.BotPass,
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(config.APIURL+"/users/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println("Login failed:", err)
		return "", ""
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ExternalID string `json:"user_external_id"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Println("Decode error:", err)
		return "", ""
	}
	return res.AccessToken, res.User.ExternalID
}
func handleSingleChat(token, botID, chatID string, ctx context.Context, pool *pgxpool.Pool) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	if _, loaded := activeChats.LoadOrStore(chatID, true); loaded {
		return
	}
	defer activeChats.Delete(chatID)

	fmt.Printf(" Bot joining chat: %s\n", chatID)
	u, _ := url.Parse(config.WS_URL + "/ws/chats/" + chatID)
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Printf("Bot WS Dial error for chat %s: %v\n", chatID, err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Bot disconnected from chat %s: %v\n", chatID, err)
			return
		}

		var msgObj WSMessage
		if err := json.Unmarshal(message, &msgObj); err == nil {
			if msgObj.SenderExternalID != botID && msgObj.Content != "" {
				fmt.Printf(" Received in %s: %s\n", chatID, msgObj.Content)

				go func(userMsg string) {
					chunks, err := retrieveFromPostgres(ctx, pool, userMsg, 3)
					if err != nil {
						fmt.Printf("Error retrieving chunks: %v\n", err)
						return
					}

					ans, err := chatCompletion(userMsg, chunks)
					if err != nil {
						fmt.Printf("Error getting completion: %v\n", err)
						return
					}

					reply := map[string]string{"content": ans}
					if b, err := json.Marshal(reply); err == nil {

						err := conn.WriteMessage(websocket.TextMessage, b)
						if err != nil {
							fmt.Printf("Error sending reply: %v\n", err)
							return
						}
						fmt.Printf(" Bot replied in %s\n", chatID)
					}
				}(msgObj.Content)
			}
		}
	}
}

func StartBot(ctx context.Context, pool *pgxpool.Pool) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	fmt.Println("Starting AI Admin Bot...")
	registerBot()
	token, botID := loginBot()
	if token == "" {
		fmt.Println("Bot login failed")
		return
	}
	fmt.Println("Bot logged in successfully")

	u, _ := url.Parse(config.WS_URL + "/ws/admin/chats")
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Printf("Admin WS Dial error: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Bot listening for new chats...")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			return
		}

		var chats []ChatItem
		if err := json.Unmarshal(message, &chats); err == nil {
			for _, chat := range chats {
				if chat.Status == "open" || chat.Status == "pending" {
					go handleSingleChat(token, botID, chat.ChatExternalID, ctx, pool)
				}
			}
		}
	}
}
