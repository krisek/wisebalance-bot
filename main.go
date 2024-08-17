package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/matrix-org/gomatrix"
)

type TotalWorth struct {
	Value float64 `json:"value"`
}

type Balance struct {
	Currency   string    `json:"currency"`
	TotalWorth TotalWorth `json:"totalWorth"`
}

var (
	apiKey     string
	profileID  string
	userToken  string
	matrixURL  string
	matrixUser string
	matrixPass string
)

func init() {
	apiKey = os.Getenv("API_KEY")
	profileID = os.Getenv("PROFILE_ID")
	userToken = os.Getenv("USER_TOKEN")
	matrixURL = os.Getenv("MATRIX_URL")
	matrixUser = os.Getenv("MATRIX_USER")
	matrixPass = os.Getenv("MATRIX_PASS")

	if apiKey == "" || profileID == "" || userToken == "" || matrixURL == "" || matrixUser == "" || matrixPass == "" {
		log.Fatal("API_KEY, PROFILE_ID, USER_TOKEN, MATRIX_URL, MATRIX_USER, and MATRIX_PASS are required environment variables")
	}
}

func getFilteredBalances() (string, error) {
	url := fmt.Sprintf("https://api.wise.com/v4/profiles/%s/balances?types=STANDARD", profileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println("Raw JSON response:", string(body))

	var balances []Balance
	if err := json.Unmarshal(body, &balances); err != nil {
		return "", err
	}

	var textOutput string
	for _, balance := range balances {
		textOutput += fmt.Sprintf("Currency: %s, Total Worth: %.2f\n", balance.Currency, balance.TotalWorth.Value)
	}

	return textOutput, nil
}

func main() {
	cli, err := gomatrix.NewClient(matrixURL, "", "")
	if err != nil {
		log.Fatalf("Failed to create Matrix client: %v", err)
	}

	resp, err := cli.Login(&gomatrix.ReqLogin{
		Type:     "m.login.password",
		User:     matrixUser,
		Password: matrixPass,
	})
	if err != nil {
		log.Fatalf("Failed to login to Matrix: %v", err)
	}

	cli.SetCredentials(resp.UserID, resp.AccessToken)

	syncer := cli.Syncer.(*gomatrix.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *gomatrix.Event) {
		if ev.Sender == cli.UserID {
			return
		}

		content := ev.Content["body"].(string)
		if strings.Contains(strings.ToLower(content), "penz") || strings.Contains(strings.ToLower(content), "pénz") {
			log.Println("Keyword detected, fetching balance...")

			balanceText, err := getFilteredBalances()
			if err != nil {
				log.Printf("Failed to fetch balance: %v", err)
				return
			}

			message := gomatrix.TextMessage{
				MsgType: "m.text",
				Body:    balanceText,
			}
			if _, err := cli.SendMessageEvent(ev.RoomID, "m.room.message", message); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	})

	log.Println("Bot is running. Press Ctrl+C to exit.")
	if err := cli.Sync(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
}
