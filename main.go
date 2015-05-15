// This file is part of the Go Get My Logs program.
//
// Copyright (C) 2015 Bryan Davis and contributors
//
// This software may be modified and distributed under the terms of the MIT
// license. See the LICENSE file for details.

/*
Go Get My Logs.

Query a Logstash managed Elasticsearch index for log events.
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"os"
	"regexp"
	"time"
)

const (
	Version             = "0.0.1"
	DefaultURL          = "http://127.0.0.1:9200"
	DefaultNumResults   = "100"
	DefaultQuery        = "*"
	DefaultDuration     = "15m"
	DefaultIndexFormat  = "logstash-%Y.%m.%d"
	DefaultOutputFormat = "{@timestamp} {host} {type} {level}: {message}"
	OutputTokenRE       = "{([^}]+)}"
)

var (
	urlFlag = kingpin.Flag("url", "Server URL").Short('u').Default(DefaultURL).OverrideDefaultFromEnvar("GGML_URL").URL()

	queryArgs   = kingpin.Arg("query", "Elasticsearch query string").Strings()
	mustFlag    = kingpin.Flag("must", "Must match").Short('m').Strings()
	mustNotFlag = kingpin.Flag("must-not", "Must not match").Short('x').Strings()

	startFlag    = kingpin.Flag("start", "Oldest timestamp to match").String()
	endFlag      = kingpin.Flag("end", "Newest timestamp to match").String()
	durationFlag = kingpin.Flag("duration", "Width of timestamp window").Short('d').Default(DefaultDuration).Duration()

	tailFlag       = kingpin.Flag("tail", "Tail event stream").Short('t').Default("false").Bool()
	numResultsFlag = kingpin.Flag("num", "Number of results to fetch").Short('n').Default(DefaultNumResults).Int()

	indexFormatFlag  = kingpin.Flag("index-format", "Index name format").Default(DefaultIndexFormat).OverrideDefaultFromEnvar("GGML_INDEX_FORMAT").String()
	outputFormatFlag = kingpin.Flag("output-format", "Output format").Short('o').Default(DefaultOutputFormat).OverrideDefaultFromEnvar("GGML_OUTPUT").String()

	verboseFlag = kingpin.Flag("verbose", "Enable verbose mode").Default("false").Bool()
	debugFlag   = kingpin.Flag("debug", "Enable debug mode").Default("false").Bool()

	tokRE = regexp.MustCompile(OutputTokenRE)
)

func main() {
	// Parse arguments
	kingpin.Version(Version)
	kingpin.CommandLine.Help = "Search for logs in a Logstash Elasticsearch index."
	kingpin.Parse()

	errorLog = log.New(os.Stderr, "ERROR ", log.Ltime|log.Lshortfile)
	if *verboseFlag || *debugFlag {
		infoLog = log.New(os.Stderr, "INFO ", log.Ltime|log.Lshortfile)
	}
	if *debugFlag {
		debugLog = log.New(os.Stderr, "TRACE ", log.Ltime|log.Lshortfile)
	}

	// Connect to the Elasticsearch cluster
	logInfo("Creating client...\n")
	client, err := elastic.NewClient(
		elastic.SetURL((*urlFlag).String()),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetErrorLog(errorLog),
		elastic.SetInfoLog(infoLog),
		elastic.SetTraceLog(debugLog))
	exitIfErr(err)

	if *tailFlag {
		Tail(client)
	} else {
		Search(client)
	}
}

func Search(client *elastic.Client) {
	logInfo("Creating search query...\n")
	query, err := NewSearchQuery()
	exitIfErr(err)

	logInfo("Searching...")
	res, err := query.Search(client)
	exitIfErr(err)

	ShowResults(res)
}

func Tail(client *elastic.Client) {
	var last time.Time
	for now := range time.Tick(10 * time.Second) {
		if last.IsZero() {
			last = now.Add(-10 * time.Second).UTC()
		}
		logInfo("Creating scroll query...\n")
		query, err := NewScrollQuery(last)
		exitIfErr(err)

		for {
			logInfo("Fetching scroll results...")
			res, err := query.Scroll(client)
			if err == elastic.EOS {
				logInfo("End of scroll cursor.\n")
				break
			}
			exitIfErr(err)
			ShowResults(res)
		}

		last = now.UTC()
	}
}

func ShowResults(res *elastic.SearchResult) {
	if res.Hits != nil {
		for _, hit := range res.Hits.Hits {
			event := make(map[string]interface{})
			err := json.Unmarshal(*hit.Source, &event)
			if err == nil {
				fmt.Println(tokRE.ReplaceAllStringFunc(*outputFormatFlag,
					func(m string) string {
						parts := tokRE.FindStringSubmatch(m)
						val, ok := event[parts[1]]
						if ok {
							return fmt.Sprintf("%v", val)
						} else {
							return m
						}
					}))
			}
		}
	}
}
