package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// https://developers.google.com/sheets/api/quickstart/go

var spreadsheetId = flag.String("spreadsheet", "", "")
var range_ = flag.String("range", "A1", "")
var verbose = flag.Bool("v", false, "")

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	flag.Parse()

	if *spreadsheetId == "" {
		fmt.Println("usage: google-sheets-append <options> <values...>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// https://golang.org/doc/faq#convert_slice_of_interface
	t := flag.Args()
	s := make([]interface{}, len(t))
	for i, v := range t {
		s[i] = v
	}

	valuerange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Range:          *range_,
		Values:         [][]interface{}{s},
	}

	// "INPUT_VALUE_OPTION_UNSPECIFIED"
	// "RAW"
	// "USER_ENTERED"
	valueInputOption := "USER_ENTERED"

	resp, err := srv.Spreadsheets.Values.Append(*spreadsheetId, *range_, valuerange).ValueInputOption(valueInputOption).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		fmt.Printf("%+v\n", resp.Updates)
	}
}
