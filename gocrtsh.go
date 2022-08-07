package main

// https://gitlab.com/lu-ka/gocrtsh

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
)

// slices of the results
type results []*result

// result json struct
type result struct {
	NameValue string `json:"name_value"`
}

// if slice contains string
func contains(list []string, query string) bool {
	for _, item := range list {
		if item == query {
			return true
		}
	}

	return false
}

func worker(wg *sync.WaitGroup, jobs <-chan string, printwildcards bool) {

	defer wg.Done()

	for job := range jobs {

		// for the results
		var jsondata results

		// query the site
		response, err := http.Get("https://crt.sh/?output=json&CN=" + job)
		if err != nil {
			log.Fatal(err)
		}

		if response.StatusCode != 200 {
			fmt.Println("[query] Bad statuscode from crt.sh, check connection!")
			os.Exit(1)
		}

		// convert to a slice of structs
		err = json.NewDecoder(response.Body).Decode(&jsondata)
		if err != nil {
			log.Fatal(err)
		}

		// to print every domain once
		var printed []string

		// for each found cert
		for _, cert := range jsondata {
			// only print if not already printed
			if !contains(printed, cert.NameValue) {
				// if wildcard cert is found print only if flag --wildcards is supplied
				if strings.HasPrefix(cert.NameValue, "*.") {
					if printwildcards {
						fmt.Printf("\033[33m%s\033[0m\n", cert.NameValue)
						printed = append(printed, cert.NameValue)
					} else {
						printed = append(printed, cert.NameValue)
					}
				} else {
					fmt.Println(cert.NameValue)
					printed = append(printed, cert.NameValue)
				}

			}

		}

		// sleep so that we don't dos crt.sh
		// ADJUST THIS VALUE AT YOUR OWN RISK OF GETTING WAF'D
		time.Sleep(time.Millisecond * 500)
	}
}

func main() {

	var printwildcards bool

	flag.BoolVar(&printwildcards, "wildcards", false, "Print found wildcard certificates")
	flag.Parse()

	// check that STDIN != empty
	stdinstat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if stdinstat.Mode()&os.ModeNamedPipe == 0 {
		fmt.Println("[main] STDIN was empty")
		os.Exit(1)
	}

	// waitgroup
	wg := new(sync.WaitGroup)

	// channel for the jobs
	chanJobs := make(chan string)

	go worker(wg, chanJobs, printwildcards)
	wg.Add(1)

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		// check if supplied domain is a valid DNSName to avoid later errors
		if !govalidator.IsDNSName(stdin.Text()) {
			continue
		} else {
			if len(stdin.Text()) == 0 {
				break
			}
			chanJobs <- stdin.Text()
		}
	}

	// if there was an error with STDIN
	if err := stdin.Err(); err != nil {
		fmt.Println("[main] Error while reading STDIN")
		os.Exit(1)
	}

	close(chanJobs)
	wg.Wait()
}
