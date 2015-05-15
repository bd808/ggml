// This file is part of the Go Get My Logs program.
//
// Copyright (C) 2015 Bryan Davis and contributors
//
// This software may be modified and distributed under the terms of the MIT
// license. See the LICENSE file for details.

package main

import (
	"github.com/jehiah/go-strftime"
	"github.com/olivere/elastic"
	"math"
	"strings"
	"time"
)

type Query struct {
	start      *time.Time
	end        *time.Time
	query      string
	must       []string
	mustNot    []string
	numResults int
	scrollId   string
}

func NewBaseQuery() (*Query, error) {
	q := &Query{
		query:      strings.Join(*queryArgs, " "),
		must:       *mustFlag,
		mustNot:    *mustNotFlag,
		numResults: *numResultsFlag,
	}

	if q.query == "" {
		q.query = DefaultQuery
	}
	return q, nil
}

func NewSearchQuery() (*Query, error) {
	q, err := NewBaseQuery()
	if err != nil {
		return nil, err
	}

	// Normalize duration to a positive interval
	duration := time.Duration(math.Abs((*durationFlag).Seconds())) * time.Second
	logDebug("duration: %v\n", duration)

	// Figure out the start and end times
	if *startFlag != "" {
		t, err := ParseTime(*startFlag)
		if err != nil {
			return nil, err
		}
		q.start = t
	}
	if *endFlag != "" {
		t, err := ParseTime(*endFlag)
		if err != nil {
			return nil, err
		}
		q.end = t
	}

	switch {
	case q.start == nil:
		if q.end == nil {
			// Default end of search range is now
			t := time.Now().UTC()
			q.end = &t
		}
		// Default start of search range is end-duration
		t := q.end.Add(time.Duration(duration.Seconds()*-1) * time.Second)
		q.start = &t
	case q.end == nil:
		// Start was given but not end.
		t := q.start.Add(duration)
		q.end = &t
	}
	logDebug("start: %s; end: %s\n", q.start, q.end)

	return q, nil
}

func NewScrollQuery(since time.Time) (*Query, error) {
	q, err := NewBaseQuery()
	if err != nil {
		return nil, err
	}
	q.start = &since
	now := time.Now().UTC()
	q.end = &now
	return q, nil
}

func (q *Query) Index() string {
	// TODO: assumes that Logstash rotates index daily
	var indices []string
	end := q.end.Truncate(time.Hour*24).AddDate(0, 0, 1)
	t := q.start.Truncate(time.Hour * 24)
	for t.Before(end) {
		indices = append(indices, strftime.Format(*indexFormatFlag, t))
		t = t.AddDate(0, 0, 1)
	}
	return strings.Join(indices, ",")
}

func (q *Query) Query() elastic.Query {
	qs := elastic.NewQueryStringQuery(q.query)

	dateRange := elastic.NewRangeFilter("@timestamp").
		From(q.start.Unix() * 1000)
	if q.end != nil {
		dateRange = dateRange.To(q.end.Unix() * 1000)
	} else {
		dateRange = dateRange.To(time.Now().UTC().Unix() * 1000)
	}

	filt := elastic.NewBoolFilter().Must(dateRange)

	for _, must := range q.must {
		filt = filt.Must(elastic.NewQueryFilter(
			elastic.NewQueryStringQuery(must)).Cache(true))
	}

	for _, mustNot := range q.mustNot {
		filt = filt.MustNot(elastic.NewQueryFilter(
			elastic.NewQueryStringQuery(mustNot)).Cache(true))
	}

	return elastic.NewFilteredQuery(qs).Filter(filt)
}

func (q *Query) Search(client *elastic.Client) (*elastic.SearchResult, error) {
	return client.Search().
		Index(q.Index()).
		Query(q.Query()).
		Sort("@timestamp", true).
		From(0).Size(q.numResults).
		Pretty(true).
		Do()
}

func (q *Query) Scroll(client *elastic.Client) (*elastic.SearchResult, error) {
	if q.scrollId == "" {
		logInfo("Getting scrollId...\n")
		res, err := client.Scroll().
			Index(q.Index()).
			Query(q.Query()).
			Size(50).
			Do()
		if err != nil {
			return nil, err
		}
		q.scrollId = res.ScrollId
		logInfo("Scroll cursor <%s>: %d results\n", q.scrollId, res.Hits.TotalHits)
	}

	return client.Scroll().Index(q.Index()).ScrollId(q.scrollId).Size(50).Do()
}
