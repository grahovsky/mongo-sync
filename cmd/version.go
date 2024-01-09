package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var (
	release   = "develop"
	buildDate = "2024-01-09T05:49:04"
	gitHash   = "7e102b8"
)

func version() string {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(struct {
		Release   string
		BuildDate string
		GitHash   string
	}{
		Release:   release,
		BuildDate: buildDate,
		GitHash:   gitHash,
	}); err != nil {
		return (fmt.Sprintf("error while decode version info: %v\n", err))
	}

	return buf.String()
}
