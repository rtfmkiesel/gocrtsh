package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
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

// returns a random desktop user-agent
func randomUserAgent() string {
	desktopAll := []string{
		// chrome
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// edge
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// opera
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// chrome
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// firefox
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:107.0) Gecko/20100101 Firefox/107.0",
		// safari
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
		// edge
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// opera
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// chrome
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// firefox
		"Mozilla/5.0 (X11; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// opera
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
	}

	// init rand
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// get an random int <= len(slice)
	i := r.Intn((len(desktopAll) - 0) + 0)
	// return a string
	return desktopAll[i]
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
		request.Header.Set("User-Agent", randomUserAgent())

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
