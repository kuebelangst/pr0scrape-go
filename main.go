package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Content ... to be filled with whatever the
// websites so called API throws at you
// generated with jsonutil `apiUrl`
type Content struct {
	AtEnd   bool   `json:"atEnd"`
	AtStart bool   `json:"atStart"`
	Cache   string `json:"cache"`
	Failure string `json:"error"`
	Items   []struct {
		Created  int    `json:"created"`
		Down     int    `json:"down"`
		Flags    int    `json:"flags"`
		Fullsize string `json:"fullsize"`
		ID       int    `json:"id"`
		Image    string `json:"image"`
		Mark     int    `json:"mark"`
		Promoted int    `json:"promoted"`
		Source   string `json:"source"`
		Thumb    string `json:"thumb"`
		Up       int    `json:"up"`
		User     string `json:"user"`
	} `json:"items"`
	Qc int `json:"qc"`
	Rt int `json:"rt"`
	Ts int `json:"ts"`
}

var apiURL = "http://pr0gramm.com/api/items/get?newer="
var picURL = "http://pr0gramm.com/data/images/"

// make configurable, or use os for finding calling Users $HOME
var dumpDir string

// ugly hack, works because linux doesn't care much about the file extension
var fileExt = ".pic"

// TODO: save this to a file after every set of goroutines so we can restart the scraping
// at the last set of pictures without too much pain.
var saveStateFile string

var lastID int

func init() {
	flag.StringVar(&dumpDir, "dumpDir", "/tank/pr0gramm/", "directory to dump the files to")
	flag.Parse()
}

func main() {
	saveStateFile = dumpDir + "savestate"
	idFromFile, err := ioutil.ReadFile(saveStateFile)
	lastID, _ = strconv.Atoi(string(idFromFile))
	if err != nil {
		fmt.Println("Oh noes, can't find the last ID")
		log.Fatal(err)
	}
	lastID = scrapeIDs(lastID)
	fmt.Println("Ferdsch...")
	os.Exit(0)
}

// Fetches API output and scrapes for the download URL and passes it to fetchImage()
func scrapeIDs(startID int) int {

	resp, err := http.Get(apiURL + strconv.Itoa(lastID))
	if err != nil {
		log.Println(err)
	}

	m := &Content{}
	if resp.StatusCode == 200 {
		content, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(content, &m)
		if err != nil {
			// if this fails the API's json output has changed so we want to choke on this
			log.Fatal(err)
		}
		for i := range m.Items {
			lastID = m.Items[i].ID

			talk := make(chan int)
			go func() {
				// Can't use m.Items[i].Source here because at leat on 4chan they delete content after a while
				fetchImage(m.Items[i].Image, lastID)
				talk <- 0
			}()
			<-talk
		}
	}

	// save the lastId to a file here
	log.Printf("Wrote new ID: %d to save state file", lastID)
	err = ioutil.WriteFile(saveStateFile, []byte(strconv.Itoa(lastID)), 0644)

	// be nice to pr0gramm
	time.Sleep(time.Second * 10)

	if m.AtStart {
		return -1
	}

	return scrapeIDs(lastID)
}

// Fetches the from `path` to file `id`.pic.
// Returns a bool signalling wether it was successful.
func fetchImage(path string, ID int) bool {
	var success = false

	resp, err := http.Get(picURL + string(path))
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode == 200 {
		pic, err := ioutil.ReadAll(resp.Body)
		err = ioutil.WriteFile(dumpDir+strconv.Itoa(ID)+fileExt, pic, 0644)
		if err != nil {
			log.Println(err)
		}
		success = true
	}

	return success
}
