package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/igorwwwwwwwwwwwwwwwwwwww/google-sheets-append/auth"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// https://developers.google.com/sheets/api/quickstart/go

var range_ = flag.String("range", "A1", "")
var verbose = flag.Bool("v", false, "")

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("usage: google-sheets-append <options> <spreadsheet-id> <values...>")
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

	client := auth.GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := flag.Arg(0)

	// https://golang.org/doc/faq#convert_slice_of_interface
	t := flag.Args()[1:]
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

	resp, err := srv.Spreadsheets.Values.Append(spreadsheetId, *range_, valuerange).ValueInputOption(valueInputOption).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		fmt.Printf("%+v\n", resp.Updates)
	}
}
