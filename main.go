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
	StartDate string   `short:"s" long:"startdate" description:"Start date (YYYY-MM-DD)" default:"2000-01-01" env:"STARTDATE"`
	EndDate   string   `short:"e" long:"enddate" description:"End date (YYYY-MM-DD)" default:"2050-01-01" env:"ENDDATE"`
	Domains   []string `short:"d" long:"domain" description:"List of domains to report on" env:"DOMAINS"`
	Html      bool     `short:"w" long:"html" description:"HTML output" env:"HTML"`
	Verbose   bool     `short:"v" long:"verbose" description:"Add verbose output to stderr" env:"VERBOSE"`
}

var cliOptions Options
var parser = flags.NewParser(&cliOptions, flags.Default)

func main() {

	// Parse flags
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

	// rspamd inbound message accepted
	//
	// 2026-03-07 16:16:41 #4493(rspamd_proxy) <94f35d>; proxy; rspamd_task_write_log: id: <a18aec2f-a5e8-48d4-83aa-498a9656f8ae@ind1s06mta1781.xt.local>, qid: <3E6337624A6>, ip: 2a00:1450:4864:20::532, from: <dawnlord66@gmail.com>, (default: F (no action): [-2.91/11.00] [ARC_ALLOW(-1.00){google.com:s=arc-20240605:i=2;},DMARC_POLICY_ALLOW(-1.00){email.trivago.com;reject;},MANY_INVISIBLE_PARTS(1.00){10;},R_DKIM_ALLOW(-1.00){email.trivago.com:s=200608;s6.y.mc.salesforce.com:s=fbldkim6;},R_SPF_ALLOW(-1.00){+ip6:2a00:1450:4000::/36;},ZERO_FONT(0.20){2;},MIME_GOOD(-0.10){multipart/alternative;text/plain;},HAS_LIST_UNSUB(-0.01){},ALIAS_RESOLVED(0.00){},ASN(0.00){asn:15169, ipnet:2a00:1450::/32, country:US;},DBL_BLOCKED_OPENRESOLVER(0.00){mail-ed1-x532.google.com:helo;mail-ed1-x532.google.com:rdns;trivago.co.uk:url;ind1s06mta1781.xt.local:mid;},DKIM_TRACE(0.00){email.trivago.com:+;s6.y.mc.salesforce.com:+;},DNSWL_BLOCKED(0.00){13.111.6.134:received;},DWL_DNSWL_NONE(0.00){salesforce.com:dkim;},FORGED_RECIPIENTS(0.00){m:dawnlord66@googlemail.com;s:dawn@plord.co.uk;},FORGED_RECIPIENTS_FORWARDING(0.00){},FORGED_SENDER(0.00){newsletter@email.trivago.com;dawnlord66@gmail.com;},FORGED_SENDER_FORWARDING(0.00){},FREEMAIL_ENVFROM(0.00){gmail.com;},FREEMAIL_TO(0.00){googlemail.com;},FROM_HAS_DN(0.00){},FROM_NEQ_ENVFROM(0.00){newsletter@email.trivago.com;dawnlord66@gmail.com;},FWD_GOOGLE(0.00){dawnlord66@googlemail.com;},HAS_REPLYTO(0.00){reply-QPXUH5FWHYZERJJY5VYT364XCE.60259@email.trivago.com;},MIME_TRACE(0.00){0:+;1:+;2:~;},MISSING_XM_UA(0.00){},RBL_SPAMHAUS_BLOCKED_OPENRESOLVER(0.00){2a00:1450:4864:20::532:from;},RCPT_COUNT_ONE(0.00){1;},RCVD_COUNT_THREE(0.00){4;},RCVD_IN_DNSWL_NONE(0.00){2a00:1450:4864:20::532:from;},RCVD_TLS_LAST(0.00){},REPLYTO_DN_EQ_FROM_DN(0.00){},REPLYTO_DOM_EQ_FROM_DOM(0.00){},REPLYTO_DOM_NEQ_TO_DOM(0.00){},TAGGED_FROM(0.00){caf_=dawn=plordcouk;},TO_DN_NONE(0.00){}]), len: 108813, time: 386.214ms, dns req: 64, digest: <d2999cc9dfa241c411c06d231b9555d1>, rcpts: <dawn@plord.co.uk>, mime_rcpts: <dawnlord66@googlemail.com,>
	// 2026-03-07T16:16:41.693676+00:00 mail postfix/qmgr[1826635]: 3E6337624A6: from=<dawnlord66+caf_=dawn=plord.co.uk@googlemail.com>, size=109379, nrcpt=1 (queue active)
	// 2026-03-07T16:16:41.892137+00:00 mail postfix/lmtp[1865617]: 3E6337624A6: to=<dawn@plord.co.uk>, relay=mail.plord.co.uk[/var/run/dovecot/lmtp], delay=0.7, delays=0.5/0.03/0.02/0.15, dsn=2.0.0, status=sent (250 2.0.0 <dawn@plord.co.uk> 6nEsLGlPrGmSdxwAnshy3A Saved)

	// rspamd inbound message rejected
	//
	// 2026-03-08T06:31:31.567386+00:00 mail postfix/cleanup[1984364]: 421FD7628B7: milter-reject: END-OF-MESSAGE from unknown[84.234.31.254]: 4.7.1 Spam message rejected; from=<04bIpYadeb@hanwayfilms.com> to=<peter@plord.co.uk> proto=ESMTP helo=<quantaq.com>
	// 2026-03-07T13:58:57.292179+00:00 mail postfix/cleanup[1846368]: 86B22762900: milter-reject: END-OF-MESSAGE from mx2.adwebmasters.digital[69.30.249.19]: 4.7.1 Try again later; from=<julierobinson545@gmail.com> to=<peter@plord.co.uk> proto=ESMTP helo=<mx1.adwebmasters.digital>

	// Outbound
	//
	// date $1, time $2, id $3, from $4
	outboundFromRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})[^ ]* mail postfix/qmgr\\[[0-9]*\\]: ([0-9A-F]*): from=<([^>]*)>")
	// date $1, time $2, id $3, to $4, status $5
	outboundToRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})[^ ]* mail postfix/smtp\\[[0-9]*\\]: ([0-9A-F]*): to=<([^>]*)>.+status=([^ ]*) ")

	// date $1, time $2, id $3, original id $4
	requeueRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})[^ ]* mail postfix/pickup\\[[0-9]*\\]: ([0-9A-F]*): uid=[^ ]* from=[^ ]* orig_id=([0-9A-F]*)")

	// Inbound
	//
	// date $1, time $2, id $3, to $4, status $5
	inboundToRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})[^ ]* mail postfix/lmtp\\[[0-9]*\\]: ([0-9A-F]*): to=<([^>]*)>.+status=([^ ]*) ")
	// date $1, time $2, id $3, from $4, to $5
	inboundRejectRegex, _ := regexp.Compile("^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}:[0-9]{2}:[0-9]{2})[^ ]* mail postfix/cleanup\\[[0-9]*\\]: ([0-9A-F]*): milter-reject: .+ Spam message rejected; from=<([^>]*)> to=<([^>]*)>")

	srsRegex, _ := regexp.Compile("^SRS0=[^=]*=[^=]*=([^=]*)=([^@]*)@")

	type Key struct {
		Date, From string
	}

	// Temp store to correlate lines
	//
	outboundFromMap := make(map[string]string)
	outboundRequeueMap := make(map[string]string)
	outboundFailedMap := make(map[string]Key)
	outboundSuccessMap := make(map[string]Key)
	inboundFailedMap := make(map[string]Key)

	// Reports
	//
	outboundSentReport := make(map[Key]int)
	outboundFailedReport := make(map[Key]int)
	outboundResentReport := make(map[Key]int)
	inboundSentReport := make(map[Key]int)
	inboundFailedReport := make(map[Key]int)

	// Loop through supplied log filenames
	//
	for _, filename := range args {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var scanner *bufio.Scanner
		if strings.HasSuffix(filename, ".gz") {
			// GZ file
			//
			gr, err := gzip.NewReader(file)
			if err != nil {
				log.Fatal(err)
			}
			defer gr.Close()
			scanner = bufio.NewScanner(gr)
		} else {
			// Plain file
			//
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {

			// Record id vs from address
			//
			fromMatches := outboundFromRegex.FindStringSubmatch(scanner.Text())
			if len(fromMatches) > 0 {

				// date $1, time $2, id $3, from $4
				outboundFromId := fromMatches[3]
				outboundFromFrom := fromMatches[4]

				if len(outboundFromFrom) == 0 {
					outboundFromMap[outboundFromId] = "unknown"
				} else {
					deSrs := srsRegex.FindStringSubmatch(outboundFromFrom)
					if len(deSrs) > 0 {
						outboundFromMap[outboundFromId] = deSrs[2] + "@" + deSrs[1]
					} else {
						outboundFromMap[outboundFromId] = outboundFromFrom
					}
				}
				continue
			}

			// Process outbound
			//
			outboundToMatches := outboundToRegex.FindStringSubmatch(scanner.Text())
			if len(outboundToMatches) > 0 {

				// date $1, time $2, id $3, to $4, status $5
				outboundToDate := outboundToMatches[1]
				outboundToTime := outboundToMatches[2]
				outboundToId := outboundToMatches[3]
				outboundToTo := outboundToMatches[4]
				outboundToStatus := outboundToMatches[5]

				// Only process if we found a previous from address for this id
				//
				from, ok := outboundFromMap[outboundToId]
				if ok {

					// Skip if outside time range
					//
					timestamp, err := time.Parse(time.DateOnly, outboundToDate)
					if err != nil {
						panic(fmt.Sprintf("could not parse timestamp: %v", err))
					}
					if timestamp.Before(startDate) || timestamp.After(endDate) {
						continue
					}

					if outboundToStatus == "sent" {

						// Record success
						//
						outboundSentReport[Key{outboundToDate, from}]++
						outboundSuccessMap[outboundToId] = Key{outboundToDate, from}

						if cliOptions.Verbose {
							fmt.Fprintf(os.Stderr, "%s %s outbound success %s -> %s\n", outboundToDate, outboundToTime, from, outboundToTo)
						}
					} else {

						// Record failure, but make sure we don't record retries as multiple failures
						//
						_, ok := outboundFailedMap[outboundToId]
						if !ok {
							outboundFailedReport[Key{outboundToDate, from}]++
							outboundFailedMap[outboundToId] = Key{outboundToDate, from}
						}

						if cliOptions.Verbose {
							fmt.Fprintf(os.Stderr, "%s %s outbound failure %s -> %s\n", outboundToDate, outboundToTime, from, outboundToTo)
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "No match for %s\n", scanner.Text())
				}
				continue
			}

			// Process inbound success
			//
			inboundToMatches := inboundToRegex.FindStringSubmatch(scanner.Text())
			if len(inboundToMatches) > 0 {

				// date $1, time $2, id $3, to $4, status $5
				inboundToDate := inboundToMatches[1]
				inboundToTime := inboundToMatches[2]
				inboundToId := inboundToMatches[3]
				inboundToTo := inboundToMatches[4]
				//inboundToStatus := inboundToMatches[5]

				// Skip if outside time range
				//
				timestamp, err := time.Parse(time.DateOnly, inboundToDate)
				if err != nil {
					panic(fmt.Sprintf("could not parse timestamp: %v", err))
				}
				if timestamp.Before(startDate) || timestamp.After(endDate) {
					continue
				}
				// Only process if we found a previous from address for this id
				//
				from, ok := outboundFromMap[inboundToId]
				if ok {
					// successful inbound
					//
					// Record success
					//
					inboundSentReport[Key{inboundToDate, inboundToTo}]++

					if cliOptions.Verbose {
						fmt.Fprintf(os.Stderr, "%s %s inbound success %s -> %s\n", inboundToDate, inboundToTime, from, inboundToTo)
					}
				} else {
					fmt.Fprintf(os.Stderr, "No match for %s\n", scanner.Text())
				}
				continue
			}

			// Process inbound reject
			//
			inboundRejectMatches := inboundRejectRegex.FindStringSubmatch(scanner.Text())
			if len(inboundRejectMatches) > 0 {

				// date $1, time $2, id $3, from $4, to $5
				inboundRejectDate := inboundRejectMatches[1]
				inboundRejectTime := inboundRejectMatches[2]
				inboundRejectId := inboundRejectMatches[3]
				inboundRejectFrom := inboundRejectMatches[4]
				inboundRejectTo := inboundRejectMatches[5]

				// Skip if outside time range
				//
				timestamp, err := time.Parse(time.DateOnly, inboundRejectDate)
				if err != nil {
					panic(fmt.Sprintf("could not parse timestamp: %v", err))
				}
				if timestamp.Before(startDate) || timestamp.After(endDate) {
					continue
				}
				// Record failure, but make sure we don't record retries as multiple failures
				//
				_, ok := inboundFailedMap[inboundRejectId]
				if !ok {
					// rejected inbound
					//
					// Record success
					//
					inboundFailedReport[Key{inboundRejectDate, inboundRejectTo}]++
					inboundFailedMap[inboundRejectId] = Key{inboundRejectDate, inboundRejectTo}

					if cliOptions.Verbose {
						fmt.Fprintf(os.Stderr, "%s %s inbound reject %s -> %s\n", inboundRejectDate, inboundRejectTime, inboundRejectFrom, inboundRejectTo)
					}
				}
				continue
			}

			// Record that an email has been re-queued (has a new id)
			//
			requeueMatches := requeueRegex.FindStringSubmatch(scanner.Text())

			if len(requeueMatches) > 0 {
				// date $1, time $2, id $3, original id $4
				outboundRequeueId := requeueMatches[3]
				outboundRequeueOrigId := requeueMatches[4]

				outboundRequeueMap[outboundRequeueOrigId] = outboundRequeueId
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	// Check to see if a outbound failure was later resent
	//
	for k := range outboundFailedMap {
		_, ok := outboundSuccessMap[k]
		if ok {
			// Normal retry and now success
			//
			outboundFailedReport[outboundFailedMap[k]]--
			continue
		}

		forwardingId := k
		forwarded := false
		for {
			// can have more than one new id, so keep following requeue map until blank
			//
			key, ok := outboundSuccessMap[outboundRequeueMap[forwardingId]]
			if ok {
				outboundFailedReport[key]--
				outboundResentReport[key]++
				forwarded = true
				break
			} else {
				forwardingId, ok = outboundRequeueMap[forwardingId]
				if !ok {
					break
				}
			}
		}
		if forwarded {
			continue
		}

		// debug for manual checking logs
		//
		fmt.Fprintf(os.Stderr, "Outbound message failed delivery %s %v\n", k, outboundFailedMap[k])
	}

	// generate reports
	//
	var outboundSentKeys []Key
	for k := range outboundSentReport {
		outboundSentKeys = append(outboundSentKeys, k)
	}
	sort.Slice(outboundSentKeys, func(i, j int) bool {
		return outboundSentKeys[i].Date < outboundSentKeys[j].Date
	})
	var outboundFailedKeys []Key
	for k := range outboundFailedReport {
		outboundFailedKeys = append(outboundFailedKeys, k)
	}
	sort.Slice(outboundFailedKeys, func(i, j int) bool {
		return outboundFailedKeys[i].Date < outboundFailedKeys[j].Date
	})
	var inboundSentKeys []Key
	for k := range inboundSentReport {
		inboundSentKeys = append(inboundSentKeys, k)
	}
	sort.Slice(inboundSentKeys, func(i, j int) bool {
		return inboundSentKeys[i].Date < inboundSentKeys[j].Date
	})
	var inboundFailedKeys []Key
	for k := range inboundFailedReport {
		inboundFailedKeys = append(inboundFailedKeys, k)
	}
	sort.Slice(inboundFailedKeys, func(i, j int) bool {
		return inboundFailedKeys[i].Date < inboundFailedKeys[j].Date
	})
	if cliOptions.Html {
		fmt.Printf("<html>\n")
	}
	for _, domain := range cliOptions.Domains {
		totalOutboundSuccess := 0
		totalOutboundResent := 0
		if cliOptions.Html {
			fmt.Printf("<h2>Domain %s outbound success:</h2>\n<table style=\"border: 1px solid;\"><thead><tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Date</b></td><td style=\"border: 1px solid;\"><b>Address</b></td><td style=\"border: 1px solid;\"><b>Sent</b></td><td style=\"border: 1px solid;\"><b>Retries</b></td></tr></thead>\n<tbody>\n", domain)
		} else {
			fmt.Printf("Domain %s outbound success:\n\n", domain)
		}
		for _, sent := range outboundSentKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if outboundSentReport[sent] > 0 {
					if outboundResentReport[sent] > 0 {
						percentRetries := float64(100) * (float64(outboundResentReport[sent]) / float64(outboundSentReport[sent]))
						if cliOptions.Html {
							fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%d</td><td style=\"border: 1px solid;\">%d or %0.2f%%</td></tr>\n", sent.Date, sent.From, outboundSentReport[sent], outboundResentReport[sent], percentRetries)
						} else {
							fmt.Printf("%s %s %d (%d, or %0.2f%%, success after requeue)\n", sent.Date, sent.From, outboundSentReport[sent], outboundResentReport[sent], percentRetries)
						}
					} else {
						if cliOptions.Html {
							fmt.Printf("<tr><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%d</td><td style=\"border: 1px solid;\"></td></tr>\n", sent.Date, sent.From, outboundSentReport[sent])
						} else {
							fmt.Printf("%s %s %d\n", sent.Date, sent.From, outboundSentReport[sent])
						}
					}
					totalOutboundSuccess = totalOutboundSuccess + outboundSentReport[sent]
					totalOutboundResent = totalOutboundResent + outboundResentReport[sent]
				}
			}
		}
		if totalOutboundSuccess > 0 {
			if !cliOptions.Html {
				fmt.Printf("\n")
			}
		}
		percentRetries := float64(100) * (float64(totalOutboundResent) / float64(totalOutboundSuccess))
		if cliOptions.Html {
			fmt.Printf("<tr><td style=\"border: 1px solid;\"><b>Total</b></td><td style=\"border: 1px solid;\"></td><td style=\"border: 1px solid;\"><b>%d</b></td><td style=\"border: 1px solid;\"><b>%d or %0.2f%%</b></td></tr>\n</tbody>\n</table>\n", totalOutboundSuccess, totalOutboundResent, percentRetries)
		} else {
			fmt.Printf("Total outbound success for domain %s %d (%d, or %0.2f%%, success after requeue)\n\n", domain, totalOutboundSuccess, totalOutboundResent, percentRetries)
		}

		totalOutboundFailed := 0
		if cliOptions.Html {
			fmt.Printf("<h2>Domain %s outbound failure:</h2>\n<table style=\"border: 1px solid;\"><thead><tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Date</b></td><td style=\"border: 1px solid;\"><b>Address</b></td><td style=\"border: 1px solid;\"><b>Failed</b></td></tr></thead>\n<tbody>\n", domain)
		} else {
			fmt.Printf("Domain %s outbound failure:\n\n", domain)
		}
		for _, sent := range outboundFailedKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if outboundFailedReport[sent] > 0 {
					if cliOptions.Html {
						fmt.Printf("<tr><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%d</td></tr>\n", sent.Date, sent.From, outboundFailedReport[sent])
					} else {
						fmt.Printf("%s %s %d\n", sent.Date, sent.From, outboundFailedReport[sent])
					}
					totalOutboundFailed = totalOutboundFailed + outboundFailedReport[sent]
				}
			}
		}
		if totalOutboundFailed > 0 {
			if !cliOptions.Html {
				fmt.Printf("\n")
			}
		}
		percentOutboundFailed := float64(100) * (float64(totalOutboundFailed) / float64(totalOutboundSuccess)) /// can be NAN
		if cliOptions.Html {
			fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Total</b></td><td style=\"border: 1px solid;\"></td><td style=\"border: 1px solid;\"><b>%d or %0.2f%%</b></td></tr>\n</tbody>\n</table>\n", totalOutboundFailed, percentOutboundFailed)
		} else {
			fmt.Printf("Total outbound failure for domain %s %d or %0.2f%%\n\n", domain, totalOutboundFailed, percentOutboundFailed)
		}

		totalInboundSuccess := 0
		if cliOptions.Html {
			fmt.Printf("<h2>Domain %s inbound success:</h2>\n<table style=\"border: 1px solid;\"><thead><tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Date</b></td><td style=\"border: 1px solid;\"><b>Address</b></td><td style=\"border: 1px solid;\"><b>Received</b></td></tr></thead>\n<tbody>\n", domain)
		} else {
			fmt.Printf("Domain %s inbound success:\n\n", domain)
		}
		for _, sent := range inboundSentKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if inboundSentReport[sent] > 0 {
					if cliOptions.Html {
						fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%d</td></tr>\n", sent.Date, sent.From, inboundSentReport[sent])
					} else {
						fmt.Printf("%s %s %d\n", sent.Date, sent.From, inboundSentReport[sent])
					}
					totalInboundSuccess = totalInboundSuccess + inboundSentReport[sent]
				}
			}
		}
		if totalInboundSuccess > 0 {
			if !cliOptions.Html {
				fmt.Printf("\n")
			}
		}
		if cliOptions.Html {
			fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Total</b></td><td style=\"border: 1px solid;\"></td><td style=\"border: 1px solid;\"><b>%d</b></td></tr>\n</tbody>\n</table>\n", totalInboundSuccess)
		} else {
			fmt.Printf("Total inbound success for domain %s %d\n\n", domain, totalInboundSuccess)
		}

		totalInboundFailed := 0
		if cliOptions.Html {
			fmt.Printf("<h2>Domain %s inbound rejected:</h2>\n<table style=\"border: 1px solid;\"><thead><tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Date</b></td><td style=\"border: 1px solid;\"><b>Address</b></td><td style=\"border: 1px solid;\"><b>Rejected</b></td></tr></thead>\n<tbody>\n", domain)
		} else {
			fmt.Printf("Domain %s inbound rejected:\n\n", domain)
		}
		for _, sent := range inboundFailedKeys {
			if strings.HasSuffix(sent.From, "@"+domain) {
				if inboundFailedReport[sent] > 0 {
					if cliOptions.Html {
						fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%s</td><td style=\"border: 1px solid;\">%d</td></tr>\n", sent.Date, sent.From, inboundFailedReport[sent])
					} else {
						fmt.Printf("%s %s %d\n", sent.Date, sent.From, inboundFailedReport[sent])
					}
					totalInboundFailed = totalInboundFailed + inboundFailedReport[sent]
				}
			}
		}
		if totalInboundFailed > 0 {
			if !cliOptions.Html {
				fmt.Printf("\n")
			}
		}
		percentInboundFailed := float64(100) * (float64(totalInboundFailed) / float64(totalInboundSuccess)) // can be NAN
		if cliOptions.Html {
			fmt.Printf("<tr style=\"border: 1px solid;\"><td style=\"border: 1px solid;\"><b>Total</b></td><td style=\"border: 1px solid;\"></td><td style=\"border: 1px solid;\"><b>%d or %0.2f%%</b></td></tr>\n</tbody>\n</table>\n", totalInboundFailed, percentInboundFailed)
		} else {
			fmt.Printf("Total inbound rejected for domain %s %d or %0.2f%%\n\n", domain, totalInboundFailed, percentInboundFailed)
		}

	}
	if cliOptions.Html {
		fmt.Printf("</html>\n")
	}
}
