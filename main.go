package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/secrets"
	"github.com/sirupsen/logrus"
)

const (
	// Label to specify the Passwordstate password list ID
	listIDLabel = "password_list_id"
)

var (
	log = logrus.New()
)

type passwordStateSecretsDriver struct {
	dockerClient *client.Client
	baseURL      string
	apiKey       string
}

type passwordResponse struct {
	PasswordID int    `json:"PasswordID"`
	Title      string `json:"Title"`
	Password   string `json:"Password"`
}

type passwordHistory struct {
	PasswordID int    `json:"PasswordID"`
	Title      string `json:"Title"`
	Password   string `json:"Password"`
	ExpiryDate string `json:"ExpiryDate"`
}

func (d passwordStateSecretsDriver) Get(req secrets.Request) secrets.Response {
	log.Infof("pwdstate: Secret: %v requested", req.SecretName)

	errorResponse := func(s string, err error) secrets.Response {
		log.Errorf("Error getting secret %q: %s: %v", req.SecretName, s, err)
		return secrets.Response{
			Value: []byte("-"),
			Err:   fmt.Sprintf("%s: %v", s, err),
		}
	}
	valueResponse := func(s string) secrets.Response {
		return secrets.Response{
			Value:      []byte(s),
			DoNotReuse: true,
		}
	}

	listID, exists := req.SecretLabels[listIDLabel]
	if !exists {
		return errorResponse("Missing required label", fmt.Errorf("label %q not found", listIDLabel))
	}

	url := fmt.Sprintf("%s/searchpasswords/%s?title=%s&PreventAuditing=true", d.baseURL, listID, url.QueryEscape(req.SecretName))
	client := &http.Client{}
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errorResponse("Error creating HTTP request", err)
	}
	httpReq.Header.Set("APIKey", d.apiKey)
	resp, err := client.Do(httpReq)
	if err != nil {
		return errorResponse("Error searching for password", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errorResponse("Unexpected response from Passwordstate", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	var passwords []passwordResponse
	if err := json.NewDecoder(resp.Body).Decode(&passwords); err != nil {
		return errorResponse("Error decoding password response", err)
	}

	if len(passwords) == 0 {
		return errorResponse("Password not found", fmt.Errorf("no password found with title %q in list %q", req.SecretName, listID))
	}
	password := passwords[0]

	historyURL := fmt.Sprintf("%s/passwordhistory/%d", d.baseURL, password.PasswordID)
	httpReq, err = http.NewRequest("GET", historyURL, nil)
	if err != nil {
		return errorResponse("Error creating history request", err)
	}
	httpReq.Header.Set("APIKey", d.apiKey)
	resp, err = client.Do(httpReq)
	if err != nil {
		return errorResponse("Error fetching password history", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errorResponse("Unexpected response from Passwordstate history", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	var history []passwordHistory
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return errorResponse("Error decoding history response", err)
	}

	if len(history) == 0 {
		return errorResponse("No password history found", fmt.Errorf("no history for password ID %d", password.PasswordID))
	}

	// Find the latest valid password (not expired or expiring tomorrow)
	now := time.Now()
	var latestValid *passwordHistory
	for i, entry := range history {
		// Parse ExpiryDate (format: "YYYY-MM-DD")
		expiry, err := time.Parse("2.1.2006", entry.ExpiryDate)
		if err != nil {
			log.Warnf("Error parsing expiry date %q for password ID %d: %v", entry.ExpiryDate, entry.PasswordID, err)
			continue
		}

		// Consider password valid if it hasn't expired or isn't expiring tomorrow
		if now.Before(expiry.AddDate(0, 0, -1)) {
			latestValid = &history[i]
			break
		}
	}

	if latestValid == nil {
		return errorResponse("No valid password found", fmt.Errorf("all passwords for %q have expired or expire tomorrow", req.SecretName))
	}

	return valueResponse(latestValid.Password)
}

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	baseURL := os.Getenv("PASSWORDSTATE_BASE_URL")
	if baseURL == "" {
		log.Fatal("PASSWORDSTATE_BASE_URL environment variable is required")
	}
	apiKey := os.Getenv("PASSWORDSTATE_API_KEY")
	if apiKey == "" {
		log.Fatal("PASSWORDSTATE_API_KEY environment variable is required")
	}

	d := passwordStateSecretsDriver{
		dockerClient: cli,
		baseURL:      baseURL,
		apiKey:       apiKey,
	}
	h := secrets.NewHandler(d)

	log.Infof("pwdstate: Starting Docker secrets plugin")
	if err := h.ServeUnix("pwdstate", 0); err != nil {
		log.Errorf("Error serving pwdstate: %v", err)
	}
}
