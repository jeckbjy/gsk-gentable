package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Build(credentials string) error {
	if credentials == "" {
		credentials = "credentials.json"
	}

	b, err := ioutil.ReadFile(credentials)
	if err != nil {
		return fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	// fetch token from web
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	fmt.Printf("Please enter authorization code:")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("Unable to retrieve token from web: %v", err)
	}

	type AuthConf struct {
		ClientID     string    `json:"client_id"`
		ClientSecret string    `json:"client_secret"`
		RedirectURL  string    `json:"redirect_url"`
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		Expiry       time.Time `json:"expiry"`
	}

	auth := AuthConf{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}

	data, err := json.MarshalIndent(auth, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal auth fail:%+v", err)
	}

	if err := ioutil.WriteFile("oauth.json", data, os.ModePerm); err != nil {
		return err
	}

	fmt.Println("save to oauth.json")

	return nil
}

func writeJson(filename string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, os.ModePerm)
}
