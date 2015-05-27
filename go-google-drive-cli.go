package main

import (
	"code.google.com/p/google-api-go-client/drive/v2"
	"flag"
	"fmt"
	tm "github.com/buger/goterm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var googleServiceKeys = flag.String("service-keys", "", "Service Keys File - Check https://developers.google.com/identity/protocols/OAuth2ServiceAccount")

func main() {
	// Get command line arguments
	flag.Parse()

	clearScreen()

	// Set Client Identity
	data, err := ioutil.ReadFile(*googleServiceKeys)
	if err != nil {
		printTitle()
		tm.Print(tm.Color("Cannot read secret file", tm.RED))
		tm.Flush()
		return
	}

	config, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/drive")
	if err != nil {
		printTitle()
		tm.Print(tm.Color("Error loading access keys", tm.RED))
		tm.Flush()
		return
	}

	client := config.Client(oauth2.NoContext)

	svc, err := drive.New(client)
	if err != nil {
		panic(err)
	}
	printTitle()
	printHelp()

MAIN_LOOP:
	for {
		fmt.Print("Type an action (h for help):")

		var inputFunction string
		n, err := fmt.Scanf("%s\r\n", &inputFunction)
		if err != nil || n != 1 {
			// handle invalid input
			fmt.Println(n, err)
			return
		}

		switch inputFunction {
		case "l":
			clearScreen()
			printTitle()
			list, err := svc.Files.List().Do()
			if err != nil {
				panic(err)
			}

			printList(list)

		case "u":
			clearScreen()
			printTitle()
			fmt.Print("Enter file to upload:")
			var fileToUpload string
			n, err := fmt.Scanf("%s\r\n", &fileToUpload)
			if err != nil || n != 1 {
				// handle invalid input
				fmt.Println(n, err)
				return
			}
			filename := filepath.Base(fileToUpload)

			// Define the metadata for the file we are going to create.
			f := &drive.File{
				Title:       filename,
				Description: filename,
			}
			// Read the file data that we are going to upload.
			m, err := os.Open(fileToUpload)
			if err != nil {
				panic(err)
			}
			// Make the API request to upload metadata and file data.
			r, err := svc.Files.Insert(f).Media(m).Do()
			if err != nil {
				panic(err)
			}
			tm.Print(tm.Color("File "+filename+" uploaded! Google Drive ID: "+r.Id, tm.BLUE))

			tm.Flush()
		case "g":
			clearScreen()
			printTitle()
			list, err := svc.Files.List().Do()
			if err != nil {
				panic(err)
			}

			printList(list)

			fmt.Print("Enter file index to download:")
			var fileToDownload int
			n, err := fmt.Scanf("%d\r\n", &fileToDownload)
			if err != nil || n != 1 {
				// handle invalid input
				fmt.Println(n, err)
				return
			}

			item := list.Items[fileToDownload]

			tm.Print(tm.Color("Downloading "+item.Title+" ...", tm.GREEN))
			tm.Flush()

			f, err := svc.Files.Get(item.Id).Do()
			if err != nil {
				panic(err)
			}
			out, err := os.Create(f.Title)
			defer out.Close()

			downloadUrl := f.DownloadUrl
			resp, err := http.Get(downloadUrl)
			if err != nil {
				panic(err)
			}
			// Make sure we close the Body later
			defer resp.Body.Close()

			// Write the gzip stream to a tmp file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				panic(err)
			}
			tm.Print(tm.Color(item.Title+" downloaded!", tm.GREEN))
			tm.Flush()

		case "d":
			clearScreen()
			printTitle()
			list, err := svc.Files.List().Do()
			if err != nil {
				panic(err)
			}

			printList(list)

			fmt.Print("Enter file index to delete:")
			var fileToDelete int
			n, err := fmt.Scanf("%d\r\n", &fileToDelete)
			if err != nil || n != 1 {
				// handle invalid input
				fmt.Println(n, err)
				return
			}

			item := list.Items[fileToDelete]

			tm.Print(tm.Color(item.Title+" will be move to trash", tm.RED))
			tm.Flush()
			fmt.Print("Are you sure? n [y/n]:")
			var questionDel string
			n, err = fmt.Scanf("%s\r\n", &questionDel)
			if err != nil || n != 1 {
				// handle invalid input
				fmt.Println(n, err)
				return
			}
			if questionDel == "y" {
				err = svc.Files.Delete(item.Id).Do()
				if err != nil {
					panic(err)
				}
				tm.Print(tm.Color(item.Title+" deleted!", tm.RED))
				tm.Flush()
			} else {
				tm.Print(tm.Color(item.Title+" deleting aborted!", tm.RED))
				tm.Flush()
			}

		case "q":
			clearScreen()
			printTitle()
			tm.Println(tm.Color("Bye!", tm.CYAN))
			tm.Flush()
			break MAIN_LOOP

		case "h":
			clearScreen()
			printTitle()
			printHelp()
		}

	}
	tm.Flush()
}
func printTitle() {
	tm.Print(tm.Color("Link", tm.BLUE))
	tm.Print(tm.Color("Swiss", tm.RED))
	tm.Println(".com - Command Line Google Drive Client")
	tm.Flush()
}
func printHelp() {
	tm.Println("l\t Listing Files")
	tm.Println("u\t Upload File")
	tm.Println("g\t (Get) Download File")
	tm.Println("d\t Delete File")
	tm.Println("q\t Quit")
	tm.Println("h\t Help")
	tm.Flush()
}

func printList(list *drive.FileList) {
	listTable := tm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprintf(listTable, "ID\tTitle\n")
	for i, file := range list.Items {

		fmt.Fprintf(listTable, "%d\t%s\n", i, file.Title)

	}
	tm.Println(listTable)
	tm.Flush()
}

func clearScreen() {
	tm.Clear()
	tm.MoveCursor(1, 1)
	tm.Flush()
}
