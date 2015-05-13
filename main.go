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
)

const (
	Version             = "0.0.1"
	DefaultURL          = "http://127.0.0.1:9200"
	DefaultNumResults   = "100"
	DefaultQuery        = "*"
	DefaultDuration     = "15m"
	DefaultIndexFormat  = "logstash-%Y.%m.%d"
	DefaultOutputFormat = "{@timestamp} {host} {type} {level}: {message}"
	OutputRE            = "{([^}]+)}"
)

var (
	urlFlag = kingpin.Flag("url", "Server URL").Short('u').Default(DefaultURL).OverrideDefaultFromEnvar("GGML_URL").URL()

	queryFlag  = kingpin.Flag("query", "Elasticsearch query string").Short('q').Default(DefaultQuery).String()
	filterFlag = kingpin.Flag("filter", "Search filter").Short('f').Strings()

	startFlag    = kingpin.Flag("start", "Oldest timestamp to match").String()
	endFlag      = kingpin.Flag("end", "Newest timestamp to match").String()
	durationFlag = kingpin.Flag("duration", "Width of timestamp window").Short('d').Default(DefaultDuration).Duration()

	numResultsFlag = kingpin.Flag("num", "Number of results to fetch").Short('n').Default(DefaultNumResults).Int()

	indexFormatFlag  = kingpin.Flag("index-format", "Index name format").Default(DefaultIndexFormat).OverrideDefaultFromEnvar("GGML_INDEX_FORMAT").String()
	outputFormatFlag = kingpin.Flag("output-format", "Output format").Short('o').Default(DefaultOutputFormat).OverrideDefaultFromEnvar("GGML_OUTPUT").String()

	verboseFlag = kingpin.Flag("verbose", "Enable verbose mode").Default("false").Bool()
	debugFlag   = kingpin.Flag("debug", "Enable debug mode").Default("false").Bool()
)

func main() {
	// Parse arguments
	kingpin.Version(Version)
	kingpin.CommandLine.Help = "Search for logs in a Logstash Elasticsearch index."
	kingpin.Parse()

	errorLog = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	if *verboseFlag || *debugFlag {
		infoLog = log.New(os.Stderr, "INFO: ", log.LstdFlags)
	}
	if *debugFlag {
		debugLog = log.New(os.Stderr, "TRACE: ", log.LstdFlags)
	}

	// Connect to the Elasticsearch cluster
	logDebug("Creating client\n")
	client, err := elastic.NewClient(
		elastic.SetURL((*urlFlag).String()),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetErrorLog(errorLog),
		elastic.SetInfoLog(infoLog),
		elastic.SetTraceLog(debugLog))
	exitIfErr(err)

	logDebug("Creating query\n")
	query, err := NewQuery()
	exitIfErr(err)

	logDebug("Searching...\n")
	res, err := query.Search(client)
	exitIfErr(err)

	re, err := regexp.Compile(OutputRE)
	exitIfErr(err)

	if res.Hits != nil {
		for _, hit := range res.Hits.Hits {
			var event map[string]interface{}
			err := json.Unmarshal(*hit.Source, &event)
			if err == nil {
				fmt.Println(re.ReplaceAllStringFunc(*outputFormatFlag,
					func(m string) string {
						parts := re.FindStringSubmatch(m)
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
