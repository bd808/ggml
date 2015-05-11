// This file is part of the Go Get My Logs program.
//
// Copyright (C) 2015 Bryan Davis and contributors
//
// This software may be modified and distributed under the terms of the MIT
// license. See the LICENSE file for details.

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	errorLog *log.Logger
	infoLog  *log.Logger
	debugLog *log.Logger
)

func ParseTime(value string) (*time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		logDebug("Parsing <%s> with format <%s>\n", value, format)
		t, err := time.Parse(format, value)
		if err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("Unkown date format '%s'", value)
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %s\n", err.Error())
		os.Exit(1)
	}
}

func logError(format string, args ...interface{}) {
	if errorLog != nil {
		errorLog.Printf(format, args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if infoLog != nil {
		infoLog.Printf(format, args...)
	}
}

func logDebug(format string, args ...interface{}) {
	if debugLog != nil {
		debugLog.Printf(format, args...)
	}
}
