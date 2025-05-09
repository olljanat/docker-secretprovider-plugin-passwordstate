package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/secrets"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

type passwordStateSecretsDriver struct {
	dockerClient *client.Client
	baseURL      string
	apiKey       string
	listID       string
}

type passwordResponse struct {
	PasswordID int    `json:"PasswordID"`
	Title      string `json:"Title"`
	Password   string `json:"Password"`
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

	url := fmt.Sprintf("%s/searchpasswords/%s?title=%s&PreventAuditing=true", d.baseURL, d.listID, url.QueryEscape(req.SecretName))
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
		return errorResponse("Password not found", fmt.Errorf("no password found with title %q in list %q", req.SecretName, d.listID))
	}
	password := passwords[0]
	return valueResponse(password.Password)
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
	listID := os.Getenv("PASSWORDSTATE_LIST_ID")
	if listID == "" {
		log.Fatal("PASSWORDSTATE_LIST_ID environment variable is required")
	}

	d := passwordStateSecretsDriver{
		dockerClient: cli,
		baseURL:      baseURL,
		apiKey:       apiKey,
		listID:       listID,
	}
	h := secrets.NewHandler(d)

	log.Infof("pwdstate: Starting Docker secrets plugin")
	if err := h.ServeUnix("pwdstate", 0); err != nil {
		log.Errorf("Error serving pwdstate: %v", err)
	}
}
