package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type UserClient interface {
	UserExists(userID uuid.UUID, token string) (bool, error)
}

type HTTPUserClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPUserClient() *HTTPUserClient {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://localhost:3001"
	}

	return &HTTPUserClient{
		baseURL: userServiceURL,
		client:  &http.Client{},
	}
}

func (c *HTTPUserClient) UserExists(userID uuid.UUID, token string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/users/profile", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, errors.New("failed to verify user existence")
}
