package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type InitRequest struct {
	SecretShares    uint8 `json:"secret_shares"`
	SecretThreshold uint8 `json:"secret_threshold"`
}

type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

type Client struct {
	URL      string
	DoUnseal bool
}

func (v *Client) Init(recoveryShares, recoveryThreshold uint8) (*InitResponse, error) {
	if recoveryShares < recoveryThreshold {
		return nil, errors.New("invalid seal configuration: threshold cannot be larger than shares")
	}

	if recoveryShares > 1 && recoveryThreshold <= 1 {
		return nil, errors.New("invalid seal configuration: threshold must be greater than one for multiple shares")
	}

	client := &http.Client{}

	// payload for the init request
	payload, err := json.Marshal(InitRequest{
		SecretShares:    recoveryShares,
		SecretThreshold: recoveryThreshold,
	})

	if err != nil {
		return nil, err
	}

	initURL := fmt.Sprintf("%s/v1/sys/init", v.URL)

	// build request
	req, err := http.NewRequest(http.MethodPut, initURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// PUT request to Vault
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse response body as JSON
	var initResponse InitResponse
	if err := json.Unmarshal(body, &initResponse); err != nil {
		return nil, err
	}

	// unseal Vault if configured to do so
	if v.DoUnseal {
		if err := v.Unseal(initResponse.KeysBase64); err != nil {
			return nil, err
		}
	}

	return &initResponse, nil
}

func (v *Client) IsInitialized() bool {
	return false
}

func (v *Client) Unseal(keysBase64 []string) error {
	unsealURL := fmt.Sprintf("%s/v1/sys/unseal", v.URL)

	client := &http.Client{}

	for _, key := range keysBase64 {

		payload := []byte(fmt.Sprintf("{\"key\": \"%s\"}", key))

		req, err := http.NewRequest(http.MethodPut, unsealURL, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}

		// PUT request to Vault
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode > 400 {
			return err
		}
	}

	return nil
}

func NewVaultClient(url string) (*Client, error) {
	if url == "" {
		return nil, errors.New("vault URL is empty")
	}

	return &Client{
		URL: url,
	}, nil
}
