package main

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const maxLogLineBytes = 4096

var (
	accountField     = regexp.MustCompile(`(?:^|\s)user=([^\s]+)`)
	sessionUserField = regexp.MustCompile(`session opened for user ([^\s(]+)`)
	sourceField      = regexp.MustCompile(`(?:^|\s)rhost=([^\s]+)`)
	pamServiceField  = regexp.MustCompile(`pam_unix\(([^:()]+):`)
)

// logRecord retains only bounded, normalized fields needed for routing.
// Passwords and full raw records never enter detection effects.
type logRecord struct {
	SourceID      string
	Message       string
	At            time.Time
	Account       string
	SourceAddress string
	Service       string
}

// parseAuthRecord accepts the local rsyslog RFC3339 timestamp format and
// extracts PAM account and source fields when present. Missing metadata stays
// empty rather than being grouped into an unsafe shared key.
func parseAuthRecord(sourceID, line string) (logRecord, error) {
	if line == "" || len(line) > maxLogLineBytes || strings.ContainsRune(line, '\x00') {
		return logRecord{}, errors.New("invalid auth log record size")
	}
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return logRecord{}, errors.New("incomplete rsyslog auth record")
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return logRecord{}, err
	}
	message := parts[3]
	record := logRecord{SourceID: sourceID, Message: message, At: at}
	if match := accountField.FindStringSubmatch(message); len(match) == 2 {
		record.Account = match[1]
	}
	if record.Account == "" {
		if match := sessionUserField.FindStringSubmatch(message); len(match) == 2 {
			record.Account = match[1]
		}
	}
	if match := sourceField.FindStringSubmatch(message); len(match) == 2 {
		record.SourceAddress = match[1]
	}
	if match := pamServiceField.FindStringSubmatch(message); len(match) == 2 {
		record.Service = match[1]
	}
	if len(record.Account) > 128 || len(record.SourceAddress) > 128 || len(record.Service) > 128 {
		return logRecord{}, errors.New("auth record field too long")
	}
	return record, nil
}
