package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	resty "github.com/go-resty/resty/v2"
	"github.com/muultipla/glim/server/api/auth"
)

// AuthTokenPath - TODO comment
func AuthTokenPath() (*string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Could not get your home directory: %v\n", err)
	}

	glimPath := fmt.Sprintf("%s/.glim", homeDir)
	if _, err := os.Stat(glimPath); os.IsNotExist(err) {
		err = os.MkdirAll(glimPath, 0700)
		if err != nil {
			return nil, fmt.Errorf("Could not create .glim in your home directory: %v", err)
		}
	}

	tokenPath := fmt.Sprintf("%s/accessToken.json", glimPath)
	return &tokenPath, nil
}

// ReadCredentials - TODO comment
func ReadCredentials() *auth.Response {
	var token auth.Response

	tokenFile, err := AuthTokenPath()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	f, err := os.Open(*tokenFile)
	if err != nil {
		fmt.Printf("Could not read file containing auth token. Please, log in again.")
		os.Exit(1)
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)
	if err := json.Unmarshal(byteValue, &token); err != nil {
		fmt.Printf("Could not get credentials from stored file %v", err)
		os.Exit(1)
	}

	return &token
}

// DeleteCredentials - TODO comment
func DeleteCredentials() {
	tokenFile, err := AuthTokenPath()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	if err := os.Remove(*tokenFile); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}

// Refresh - TODO comment
func Refresh(rt string) {
	// Rest API authentication
	client := resty.New()

	// Set bearer token
	client.SetAuthToken(rt)
	// TODO - We should verify server's certificate
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Query refresh token
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(RefreshToken{
			Token: rt,
		}).
		SetError(&APIError{}).
		Post(fmt.Sprintf("%s/login/refresh_token", url))

	if err != nil {
		fmt.Printf("Error connecting with Glim: %v\n", err)
		os.Exit(1)
	}

	if resp.IsError() {
		fmt.Printf("Error response from Glim: %v\n", resp.Error().(*APIError).Message)
		os.Exit(1)
	}

	// Authenticated, let's store tokens in $HOME/.glim/accessToken.json
	tokenFile, err := AuthTokenPath()
	if err != nil {
		fmt.Printf("Could not guess auth token path: %v", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(*tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Could not create file to store auth token: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(resp.String()); err != nil {
		fmt.Printf("Could not store credentials in our local fs: %v\n", err)
		os.Exit(1)
	}
}

// NeedsRefresh - TODO comment
func NeedsRefresh(token *auth.Response) bool {
	// Check expiration
	now := time.Now()
	expiration := time.Unix(token.ExpiresOn, 0)
	return expiration.Before(now)
}

// AmIManager - TODO comment
func AmIManager(token *auth.Response) bool {
	claims := make(jwt.MapClaims)
	jwt.ParseWithClaims(token.AccessToken, claims, nil)

	// Extract access token jti
	manager, ok := claims["manager"].(bool)
	if !ok {
		fmt.Printf("Could not parse access token. Please try to log in again\n")
		os.Exit(1)
	}

	return manager
}

// WhichIsMyTokenUID - TODO comment
func WhichIsMyTokenUID(token *auth.Response) float64 {
	claims := make(jwt.MapClaims)
	jwt.ParseWithClaims(token.AccessToken, claims, nil)

	// Extract access token jti
	uid, ok := claims["uid"].(float64)
	if !ok {
		fmt.Printf("Could not parse access token. Please try to log in again\n")
		os.Exit(1)
	}

	return uid
}
