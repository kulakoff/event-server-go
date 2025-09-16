package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SyslogMessage struct {
	Format    string
	Priority  int
	Version   int
	Timestamp string
	Hostname  string
	App       string
	PID       string
	MsgID     string
	Message   string
}

var (
	regexIETF    = regexp.MustCompile(`<(\d{1,3})>(\d+) (\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?\w?(?:[+-]\d{2}:\d{2})?) (\S+) (\S+) (\S+) (\S+)\s-\s(.*)$`)
	regexBSD     = regexp.MustCompile(`<(\d{1,3})>(\w+\s+\d{1,2}\s\d{2}:\d{2}:\d{2})\s(\S+)?\s([\w\s.]+)\s(\S+):\s(.*)$`)
	regexRubetek = regexp.MustCompile(`<(\d{1,3})>(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}) (\S+) (\S+) (.*)$`)
)

// ParseSyslogMessage parses syslog_custom messages based on the given format
func ParseSyslogMessage(str string, unit string) *SyslogMessage {
	fmt.Printf(str)
	if str == "" {
		return nil
	}

	str = strings.TrimSpace(str)

	if unit == "SERVICE_UFANET" {
		index := strings.Index(str, ": ")
		if index != -1 {
			message := str[index+2:]
			return &SyslogMessage{
				Hostname: "",
				Message:  message,
			}
		}
	}

	// Check if the message follows the RFC 5424 format
	if parts := regexIETF.FindStringSubmatch(str); parts != nil {
		fmt.Println("IETF")
		return &SyslogMessage{
			Format:    "RFC5424",
			Priority:  toInt(parts[1]),
			Version:   toInt(parts[2]),
			Timestamp: parseTimestamp(parts[3]),
			Hostname:  parts[4],
			App:       parts[5],
			PID:       parts[6],
			MsgID:     parts[7],
			Message:   parts[8],
		}
	}

	// BSD RFC 3164 format
	if parts := regexBSD.FindStringSubmatch(str); parts != nil {
		return &SyslogMessage{
			Format:    "BSD",
			Priority:  toInt(parts[1]),
			Timestamp: parseTimestamp(parts[2]),
			Hostname:  parts[3],
			App:       parts[4],
			PID:       parts[5],
			Message:   parts[6],
		}
	}

	// Rubetek format
	if parts := regexRubetek.FindStringSubmatch(str); parts != nil {
		return &SyslogMessage{
			Format:    "Rubetek",
			Priority:  toInt(parts[1]),
			Timestamp: parseTimestamp(parts[2]),
			Hostname:  parts[3],
			App:       parts[4],
			Message:   parts[5],
		}
	}

	return nil
}

// parseTimestamp parses the timestamp string to a standard format
func parseTimestamp(timestamp string) string {
	// Adjust timestamp parsing as needed
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp // fallback to original if parsing fails
	}
	return t.Format(time.RFC3339)
}

// toInt converts a string to an integer
func toInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}
