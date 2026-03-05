package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	StartDate string `short:"s" long:"startdate" description:"Start date (YYYY-MM-DD)" default:"2000-01-01" env:"STARTDATE"`
	EndDate   string `short:"a" long:"enddate" description:"End date (YYYY-MM-DD)" default:"2050-01-01" env:"ENDDATE"`
}

var cliOptions Options
var parser = flags.NewParser(&cliOptions, flags.Default)

func main() {

	// parse flags
	//
	args, err := parser.Parse()
	if err != nil {
		panic(fmt.Sprintf("could not parse cli: %v", err))
	}

	startDate, err := time.Parse(time.DateOnly, cliOptions.StartDate)
	if err != nil {
		panic(fmt.Sprintf("could not parse start time: %v", err))
	}
	endDate, err := time.Parse(time.DateOnly, cliOptions.EndDate)
	if err != nil {
		panic(fmt.Sprintf("could not parse start time: %v", err))
	}

	// supplies from & id
	//
	// 2026-03-03T10:45:29.294394+00:00 mail postfix/qmgr[470834]: 3172D762A5D: from=<leadingedge@u3acommunities.org>, size=49555, nrcpt=1 (queue active)

	// supplies to, id and status
	//
	// 2026-03-03T10:45:35.473746+00:00 mail postfix/smtp[1013986]: 3172D762A5D: to=<arharradine@me.com>, relay=mx01.mail.icloud.com[17.56.9.31]:25, delay=6.3, delays=0.1/0/1.1/5.1, dsn=2.0.0, status=sent (250 2.0.0 Ok: queued as D04B3C0008B)

	// sent via smtp2go
	//
	// 2026-03-03T10:18:23.504367+00:00 mail postfix/smtp[1010314]: A89EB7629C3: to=<bob@ashby.net>, relay=mx.netidentity.com.cust.hostedemail.com[216.40.42.4]:25, delay=0.8, delays=0.06/0/0.74/0, dsn=4.0.0, status=deferred (host mx.netidentity.com.cust.hostedemail.com[216.40.42.4] refused to talk to me: 421 Service not available, closing transmission channel)
	// 2026-03-03T11:15:05.609802+00:00 mail postfix/pickup[1018383]: 94D3C762A0C: uid=105 from=<leadingedge@u3acommunities.org> orig_id=A89EB7629C3
	// 2026-03-03T11:15:08.773836+00:00 mail postfix/smtp[1018396]: 94D3C762A0C: to=<bob@ashby.net>, relay=mail.smtp2go.com[45.79.71.155]:587, delay=3406, delays=3403/0.1/1.6/1.5, dsn=2.0.0, status=sent (250 OK id=1vxNiW-4o5NDgrjpxK-skOz)

	// date $2 - id $3 = from
	fromRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})[^ ]* mail postfix/qmgr\\[[0-9]*\\]: ([0-9A-F]*): from=<([^>]*)>")

	// date $2 - id $3 = from $4 = status
	toRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})[^ ]* mail postfix/[sl]mtp\\[[0-9]*\\]: ([0-9A-F]*): to=<([^>]*)>.+status=([^ ]*) ")

	// date $2 - new id $3 = from $4 = old id
	forwardRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})[^ ]* mail postfix/pickup\\[[0-9]*\\]: ([0-9A-F]*): uid=[^ ]* from=[^ ]* orig_id=([0-9A-F]*)")

	type Key struct {
		Date, From string
	}

	// temp store to correlate lines
	//-
	fromMap := make(map[string]string)
	forwardMap := make(map[string]string)
	failedMap := make(map[string]Key)
	successMap := make(map[string]Key)

	// reports
	//
	sentReport := make(map[Key]int)
	failedReport := make(map[Key]int)
	resentReport := make(map[Key]int)

	for _, filename := range args {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var scanner *bufio.Scanner
		if strings.HasSuffix(filename, ".gz") {
			gr, err := gzip.NewReader(file)
			if err != nil {
				log.Fatal(err)
			}
			defer gr.Close()
			scanner = bufio.NewScanner(gr)
		} else {
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {

			fromMatches := fromRegex.FindStringSubmatch(scanner.Text())
			if len(fromMatches) > 0 {
				//fmt.Fprintf(os.Stderr, "from date=%s id=%s from=%s\n", fromMatches[1], fromMatches[2], fromMatches[3])
				if len(fromMatches[3]) == 0 {
					fromMap[fromMatches[2]] = "unknown"
				} else {
					fromMap[fromMatches[2]] = fromMatches[3]
				}
				continue
			}
			toMatches := toRegex.FindStringSubmatch(scanner.Text())
			if len(toMatches) > 0 {
				from := fromMap[toMatches[2]]
				if len(from) > 0 {
					timestamp, err := time.Parse(time.DateOnly, toMatches[1])
					if err != nil {
						panic(fmt.Sprintf("could not parse timestamp: %v", err))
					}
					if timestamp.Before(startDate) || timestamp.After(endDate) {
						continue
					}
					//fmt.Fprintf(os.Stderr, "to date=%s id=%s to=%s status=%s from=%s\n", toMatches[1], toMatches[2], toMatches[3], toMatches[4], from)
					if toMatches[4] == "sent" {
						sentReport[Key{toMatches[1], from}]++
						successMap[toMatches[2]] = Key{toMatches[1], from}
					} else {
						// make sure we don't record re-tries as multiple failures
						_, ok := failedMap[toMatches[2]]
						if !ok {
							failedReport[Key{toMatches[1], from}]++
							failedMap[toMatches[2]] = Key{toMatches[1], from}
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "No match for %s\n", scanner.Text())
				}
				continue
			}
			forwardMatches := forwardRegex.FindStringSubmatch(scanner.Text())
			if len(forwardMatches) > 0 {
				//fmt.Fprintf(os.Stderr, "forward date=%s id=%s oldid=%s\n", forwardMatches[1], forwardMatches[2], forwardMatches[3])
				forwardMap[forwardMatches[3]] = forwardMatches[2]
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	// check to see if a failure was later re-sent
	//
	for k := range failedMap {
		_, ok := successMap[k]
		if ok {
			//fmt.Fprintf(os.Stderr, "  later succeeded\n")
			failedReport[failedMap[k]]--
			continue
		}

		if len(forwardMap[k]) > 0 {
			key, ok := successMap[forwardMap[k]]
			if ok {
				//fmt.Fprintf(os.Stderr, "    was successfully sent %v %v\n", key, failedReport[key])
				failedReport[key]--
				resentReport[key]++
				continue
			}
		}

		fmt.Fprintf(os.Stderr, "failed %s %v\n", k, failedMap[k])
	}

	// generate reports
	//
	var sentKeys []Key
	for k := range sentReport {
		sentKeys = append(sentKeys, k)
	}
	sort.Slice(sentKeys, func(i, j int) bool {
		return sentKeys[i].Date < sentKeys[j].Date
	})
	var failedKeys []Key
	for k := range failedReport {
		failedKeys = append(failedKeys, k)
	}
	sort.Slice(failedKeys, func(i, j int) bool {
		return failedKeys[i].Date < failedKeys[j].Date
	})
	for _, domain := range []string{"u3acommunities.org", "u3a.social", "plord.co.uk"} {
		total := 0
		fmt.Printf("Domain %s success\n\n", domain)
		for _, sent := range sentKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if sentReport[sent] > 0 {
					if resentReport[sent] > 0 {
						fmt.Printf("%s %s %d (%d via smtp2go)\n", sent.Date, sent.From, sentReport[sent], resentReport[sent])
					} else {
						fmt.Printf("%s %s %d\n", sent.Date, sent.From, sentReport[sent])
					}
					total = total + sentReport[sent]
				}
			}
		}
		fmt.Printf("\nTotal success for domain %s %d\n\n", domain, total)

		total = 0
		fmt.Printf("Domain %s failure\n\n", domain)
		for _, sent := range failedKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if failedReport[sent] > 0 {
					fmt.Printf("%s %s %d\n", sent.Date, sent.From, failedReport[sent])
					total = total + failedReport[sent]
				}
			}
		}
		fmt.Printf("\nTotal failure for domain %s %d\n\n", domain, total)
	}
}
