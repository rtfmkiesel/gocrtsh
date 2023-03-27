package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"gitlab.com/rtfmkiesel/gocrtsh/pkg/randomUserAgent"
)

// struct for the JSON response of crtsh
type jsonResponse struct {
	NameValue string `json:"name_value"`
}

// returns true if []string slice contains string
func contains(list []string, query string) bool {
	for _, item := range list {
		if item == query {
			return true
		}
	}

	return false
}

// go function to make a GET request to crt.sh's JSON API
func crtshRunner(wg *sync.WaitGroup, chanJobs <-chan string, chanResults chan<- string) {
	defer wg.Done()

	// setup http client
	client := &http.Client{}

	// for each job
	for job := range chanJobs {
		// build request
		request, err := http.NewRequest("GET", "https://crt.sh/?output=json&CN="+job, nil)
		if err != nil {
			log.Fatal(err)
		}

		// set a random User Agent
		request.Header.Set("User-Agent", randomUserAgent.Desktop())

		// make the request
		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		// break if the status code was not 200
		if response.StatusCode != 200 {
			log.Fatalf(
				"Got statuscode '%d' from crt.sh while requesting %s\n",
				response.StatusCode,
				job,
			)
		}

		// decode json response
		var results []*jsonResponse
		err = json.NewDecoder(response.Body).Decode(&results)
		if err != nil {
			log.Fatal(err)
		}

		// add the found certs to the result channel
		for _, cert := range results {
			chanResults <- cert.NameValue
		}

		// sleep so that we don't DOS crt.sh
		time.Sleep(time.Millisecond * 1000)
	}
}

// go function for printing the results
func resultsRunner(wg *sync.WaitGroup, chanResults <-chan string, printWildcards bool) {
	defer wg.Done()

	// setup a slice of already printed cert to only print uniq certs
	var printed []string

	// for each result
	for result := range chanResults {
		// cert has not been printed yet
		if !contains(printed, result) {
			// if the cert starts with a * it is a wildcard cert
			if strings.HasPrefix(result, "*.") {
				// if the user want to print wildcard certs
				if printWildcards {
					// print & append to printed
					fmt.Println(result)
					printed = append(printed, result)
				} else {
					// do not print & append
					printed = append(printed, result)
				}
			} else {
				// print & append to printed
				fmt.Println(result)
				printed = append(printed, result)
			}
		}
	}
}

func main() {
	// setup & parse args
	var flagWildCards bool
	var flagRunner int
	flag.BoolVar(&flagWildCards, "w", false, "Print wildcard certificates")
	flag.IntVar(&flagRunner, "r", 1, "Number of runners")
	flag.Parse()

	// check that stdin != empty
	stdinstat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if stdinstat.Mode()&os.ModeNamedPipe == 0 {
		log.Fatal("stdin was empty")
	}

	// setup waitgroup for the jobs
	wgCrtsh := new(sync.WaitGroup)
	// setup channel for the jobs
	chanJobs := make(chan string)

	// setup waitgroup for the output runner
	wgResults := new(sync.WaitGroup)
	// setup channel for the output runner
	chanResults := make(chan string)

	// start the crtsh runners
	for i := 0; i <= flagRunner; i++ {
		go crtshRunner(wgCrtsh, chanJobs, chanResults)
		wgCrtsh.Add(1)
	}

	// start the output runner
	go resultsRunner(wgResults, chanResults, flagWildCards)
	wgResults.Add(1)

	// setup scanner for stdin
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		// read from the scanner
		domain := stdin.Text()
		// check if supplied domain is a valid domain
		if !govalidator.IsDNSName(domain) {
			continue
		} else {
			// add a job to the channel
			chanJobs <- domain
		}
	}

	// if there was an error with STDIN
	if err := stdin.Err(); err != nil {
		log.Fatal(err)
	}

	// closing of the channel will start the crtsh runners
	close(chanJobs)
	wgCrtsh.Wait()

	// closing of the channel will start the output runner
	close(chanResults)
	wgResults.Wait()
}
