package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/projectdiscovery/retryabledns"
)

var (
	// command line arguments
	flagWildCards bool
	flagOnline    bool
	flagSilent    bool
	flagRunner    int
	// DNS resolvers for the lookup with '-o'
	dnsServers = []string{"9.9.9.9:53", "1.1.1.1:53", "8.8.8.8:53"}
)

// Struct for the JSON response of crtsh
type jsonResponse struct {
	NameValue string `json:"name_value"`
}

// catch() will handle errors
func catch(err error) {
	if !flagSilent {
		fmt.Printf("ERROR: %s\n", err)
	}
}

// catchCrit() will handle critical errors
func catchCrit(err error) {
	if !flagSilent {
		fmt.Printf("CRITICAL: %s\n", err)
	}
	os.Exit(1)
}

// contains() will return true if a []string contains a specified string
func contains(list []string, query string) bool {
	for _, item := range list {
		if item == query {
			return true
		}
	}

	return false
}

// randomUserAgent() will return a random desktop user-agent
func randomUserAgent() string {
	desktopAll := []string{
		// Chrome Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Edge Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// Opera Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// Chrome Mac
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox Mac
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.0; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Safari Mac
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
		// Edge Mac
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/107.0.1418.62",
		// Opera Mac
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
		// Chrome Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// Firefox Linux
		"Mozilla/5.0 (X11; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0",
		// Opera Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/93.0.4585.21",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Get an random int <= len(slice)
	i := r.Intn((len(desktopAll) - 0) + 0)

	return desktopAll[i]
}

// crtshRunner() is a go func to make requests to crt.sh's JSON API
func crtshRunner(wg *sync.WaitGroup, chanJobs <-chan string, chanResults chan<- string) {
	defer wg.Done()

	// Setup HTTP client
	client := &http.Client{}

	// For each job
	for job := range chanJobs {
		// Build request
		request, err := http.NewRequest("GET", "https://crt.sh/?output=json&CN="+job, nil)
		if err != nil {
			catch(err)
			continue
		}

		// Set a random user-agent
		request.Header.Set("User-Agent", randomUserAgent())

		// Make the request
		response, err := client.Do(request)
		if err != nil {
			catch(err)
			continue
		}

		// Break if the status code was not 200
		if response.StatusCode != 200 {
			catch(fmt.Errorf(
				"got statuscode '%d' from crt.sh while requesting %s",
				response.StatusCode,
				job,
			))
			continue
		}

		// Decode JSON response
		var results []*jsonResponse
		err = json.NewDecoder(response.Body).Decode(&results)
		if err != nil {
			catch(err)
			continue
		}

		// Add the found certs to the result channel
		for _, cert := range results {
			chanResults <- cert.NameValue
		}

		// Sleep so that we don't DOS crt.sh
		time.Sleep(time.Second * 1)
	}
}

// resultsRunner() is a go function for handling results
func resultsRunner(wg *sync.WaitGroup, chanResults <-chan string, printWildcards bool) {
	defer wg.Done()

	// Setup a slice of already printed cert
	// so certs will only get printed once
	var printed []string

	// Init DNS client
	dnsClient, err := retryabledns.New(dnsServers, 3)
	if err != nil {
		catchCrit(err)
	}

	// For each result
	for result := range chanResults {
		// Cert has been printed
		if contains(printed, result) {
			continue
		}

		printed = append(printed, result)

		// If the cert starts with a * it's a wildcard cert
		if strings.HasPrefix(result, "*.") {
			if !printWildcards {
				continue
			}
		}

		// Make a DNS query if specified
		if flagOnline {
			addrs, err := dnsClient.Resolve(result)
			if err != nil {
				catch(err)
				continue
			}

			// Get all the relevant records from addrs
			// Pretty ugly I know but addrs.AllRecords contains stuff we don't want
			var records []string
			records = append(records, addrs.A...)
			records = append(records, addrs.AAAA...)
			records = append(records, addrs.CNAME...)
			records = append(records, addrs.MX...)
			records = append(records, addrs.SRV...)
			records = append(records, addrs.TXT...)

			// Check if a addr was resolved
			for _, addr := range records {
				if addr != "" {
					// Print on the first record returned by the DNS query
					fmt.Println(result)
					break
				}
			}

		} else {
			// Just print
			fmt.Println(result)
		}
	}
}

func main() {
	// Parse args
	flag.BoolVar(&flagWildCards, "w", false, "")
	flag.BoolVar(&flagOnline, "o", false, "")
	flag.BoolVar(&flagSilent, "s", false, "")
	flag.IntVar(&flagRunner, "r", 1, "")
	flag.Usage = func() {
		fmt.Printf(`Usage: cat domains.txt | gocrtsh [OPTIONS]

Options:
    -w Print found wildcard certificates      (default: false)
    -o Only print online (resolvable) domains (default: false)
    -r How many runners/threads to spawn      (default: 1)
    -s Do not print errors                    (default: false)
    -h Prints this text
		`)
	}
	flag.Parse()

	// Check that stdin != empty
	stdinstat, err := os.Stdin.Stat()
	if err != nil {
		catchCrit(err)
	}
	if stdinstat.Mode()&os.ModeNamedPipe == 0 {
		catchCrit(fmt.Errorf("stdin was empty"))
	}

	// Setup waitgroup for the jobs
	wgCrtsh := new(sync.WaitGroup)
	// Setup channel for the jobs
	chanJobs := make(chan string)

	// Setup waitgroup for the output runner
	wgResults := new(sync.WaitGroup)
	// Setup channel for the output runner
	chanResults := make(chan string)

	// Start the crtsh runners
	for i := 0; i <= flagRunner; i++ {
		go crtshRunner(wgCrtsh, chanJobs, chanResults)
		wgCrtsh.Add(1)
	}

	// Start the output runner
	go resultsRunner(wgResults, chanResults, flagWildCards)
	wgResults.Add(1)

	// Setup scanner for stdin
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		// Read from the scanner
		domain := stdin.Text()
		// Check if supplied domain is a valid domain
		if !govalidator.IsDNSName(domain) {
			continue
		} else {
			// Add a job to the channel
			chanJobs <- domain
		}
	}

	// If there was an error with STDIN
	if err := stdin.Err(); err != nil {
		catchCrit(err)
	}

	// Closing the channel will start the crtsh runners
	close(chanJobs)
	wgCrtsh.Wait()

	close(chanResults)
	wgResults.Wait()
}
