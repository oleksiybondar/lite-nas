package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	maxLogLineBytes = 4096
	// maxDatagramBytes bounds one JSON-encoded record. JSON escaping can
	// expand a 4 KiB message, so the datagram limit is larger than the line.
	maxDatagramBytes = 16 * 1024
)

var (
	accountField    = regexp.MustCompile(`(?:^|\s)user=([^\s]+)`)
	sourceField     = regexp.MustCompile(`(?:^|\s)rhost=([^\s]+)`)
	pamServiceField = regexp.MustCompile(`pam_unix\(([^:()]+):`)
	// sshSourceField reads the peer of OpenSSH "from <ip> port" and
	// "user <name> <ip> port" records, which carry no PAM rhost field.
	sshSourceField = regexp.MustCompile(`(?:from|user \S*) ([0-9A-Fa-f:.]{2,45}) port [0-9]{1,5}`)
	// fallbackAccountFields are tried in order when no PAM user= field exists:
	// PAM sessions, sudo policy records, shadow-utils group, account and
	// password changes.
	fallbackAccountFields = []*regexp.Regexp{
		regexp.MustCompile(`session opened for user ([^\s(]+)`),
		regexp.MustCompile(`^([a-z_][a-z0-9_.-]{0,31}) : `),
		regexp.MustCompile(`^add '([^'\s]+)' to (?:shadow )?group `),
		regexp.MustCompile(`^user (\S+) added by \S+ to (?:shadow )?group `),
		regexp.MustCompile(`^new user: name=([^,\s]+),`),
		regexp.MustCompile(`password changed for (\S+)$`),
	}
	usbPortField = regexp.MustCompile(`^usb ([0-9]{1,3}-[0-9.]{1,32}):`)
	usbIDField   = regexp.MustCompile(`idVendor=([0-9a-f]{4}), idProduct=([0-9a-f]{4})`)
)

// rsyslogRecord is the transport contract produced by the
// LiteNASLogAnalysisRecord template in
// configs/etc/rsyslog.d/30-lite-nas-log-analysis.conf. Field names and bounds
// must stay aligned with that template.
type rsyslogRecord struct {
	Timestamp string `json:"timestamp" validate:"required,max=64"`
	Facility  string `json:"facility" validate:"required,max=16,alphanum"`
	Input     string `json:"input" validate:"required,max=32,alphanum"`
	Program   string `json:"program" validate:"max=128"`
	Message   string `json:"message" validate:"required,max=4096"`
}

// logRecord retains only bounded, normalized fields needed for routing.
// Passwords and full raw records never enter detection effects.
type logRecord struct {
	SourceID      string
	Message       string
	At            time.Time
	Account       string
	SourceAddress string
	Service       string
	Device        string
}

// decodeRsyslogRecord validates one datagram against the rsyslog template
// contract. Unknown fields, trailing JSON, NUL bytes and oversized values are
// rejected before any source selection or rule matching happens.
func decodeRsyslogRecord(data []byte, validate *validator.Validate) (rsyslogRecord, time.Time, error) {
	if len(data) == 0 || len(data) > maxDatagramBytes {
		return rsyslogRecord{}, time.Time{}, errors.New("invalid rsyslog datagram size")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var record rsyslogRecord
	if err := decoder.Decode(&record); err != nil {
		return rsyslogRecord{}, time.Time{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return rsyslogRecord{}, time.Time{}, errors.New("rsyslog datagram contains trailing data")
	}
	// rsyslog keeps the separator space that follows the syslog tag.
	record.Message = strings.TrimLeft(record.Message, " ")
	if err := validate.Struct(record); err != nil {
		return rsyslogRecord{}, time.Time{}, err
	}
	// Validator length tags count runes; the line bound is enforced in bytes.
	if len(record.Message) > maxLogLineBytes || strings.ContainsRune(record.Message, '\x00') || strings.ContainsRune(record.Program, '\x00') {
		return rsyslogRecord{}, time.Time{}, errors.New("rsyslog record message is oversized or contains NUL bytes")
	}
	at, err := time.Parse(time.RFC3339Nano, record.Timestamp)
	if err != nil {
		return rsyslogRecord{}, time.Time{}, err
	}
	return record, at, nil
}

// newLogRecord normalizes one validated record for a selecting source. The
// source format decides which bounded correlation fields are extracted;
// missing metadata stays empty rather than being grouped into a shared key.
func newLogRecord(source sourceSpec, record rsyslogRecord, at time.Time) (logRecord, error) {
	normalized := logRecord{SourceID: source.ID, Message: record.Message, At: at}
	switch source.Format {
	case formatAuth:
		extractAuthFields(&normalized)
	case formatKernel:
		extractKernelFields(&normalized)
	default:
		return logRecord{}, errors.New("unsupported source format")
	}
	if len(normalized.Account) > 128 || len(normalized.SourceAddress) > 128 || len(normalized.Service) > 128 || len(normalized.Device) > 64 {
		return logRecord{}, errors.New("log record field too long")
	}
	return normalized, nil
}

// extractAuthFields reads account, remote host and PAM service fields.
func extractAuthFields(record *logRecord) {
	if match := accountField.FindStringSubmatch(record.Message); len(match) == 2 {
		record.Account = match[1]
	}
	for _, field := range fallbackAccountFields {
		if record.Account != "" {
			break
		}
		if match := field.FindStringSubmatch(record.Message); len(match) == 2 {
			record.Account = match[1]
		}
	}
	if match := sourceField.FindStringSubmatch(record.Message); len(match) == 2 {
		record.SourceAddress = match[1]
	}
	if record.SourceAddress == "" {
		if match := sshSourceField.FindStringSubmatch(record.Message); len(match) == 2 {
			record.SourceAddress = match[1]
		}
	}
	if match := pamServiceField.FindStringSubmatch(record.Message); len(match) == 2 {
		record.Service = match[1]
	}
}

// extractKernelFields reads the USB port and hexadecimal vendor/product IDs.
// Device-reported strings such as product names are attacker controlled and
// are deliberately not copied into alert text.
func extractKernelFields(record *logRecord) {
	match := usbPortField.FindStringSubmatch(record.Message)
	if len(match) != 2 {
		return
	}
	record.Device = match[1]
	if ids := usbIDField.FindStringSubmatch(record.Message); len(ids) == 3 {
		record.Device += " " + ids[1] + ":" + ids[2]
	}
}
