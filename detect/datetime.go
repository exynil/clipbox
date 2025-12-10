package detect

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import "regexp"

var dateTimeRegexes = []*regexp.Regexp{
	// ISO 8601: 2024-01-15T10:30:00Z, 2024-01-15T10:30:00+05:00
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})$`),
	// ISO 8601 date: 2024-01-15
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
	// Common formats: 01/15/2024, 15/01/2024, 01-15-2024
	regexp.MustCompile(`^\d{1,2}[/-]\d{1,2}[/-]\d{4}$`),
	// Date with time: 01/15/2024 10:30, 01-15-2024 10:30:00
	regexp.MustCompile(`^\d{1,2}[/-]\d{1,2}[/-]\d{4}\s+\d{1,2}:\d{2}(:\d{2})?$`),
	// Date with time and AM/PM: 03/07/2025 09:15 AM, 01/15/2024 10:30 PM
	regexp.MustCompile(`^\d{1,2}[/-]\d{1,2}[/-]\d{4}\s+\d{1,2}:\d{2}(:\d{2})?\s+(AM|PM|am|pm)$`),
	// Time formats: 10:30:00, 10:30
	regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2})?$`),
	// Time with AM/PM: 09:15 AM, 10:30 PM
	regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2})?\s+(AM|PM|am|pm)$`),
	// RFC3339-like: 2024-01-15 10:30:00
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{1,2}:\d{2}:\d{2}$`),
	// Date with time: 2024-01-15 10:30:00 AM
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{1,2}:\d{2}:\d{2}\s+(AM|PM|am|pm)$`),
	// DD.MM.YYYY formats: 15.01.2024, 15.01.2024 10:30
	regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{4}(\s+\d{1,2}:\d{2}(:\d{2})?)?$`),
	// DD.MM.YYYY with AM/PM: 15.01.2024 10:30 AM
	regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{4}\s+\d{1,2}:\d{2}(:\d{2})?\s+(AM|PM|am|pm)$`),
	// DD MMM YYYY: 15 Jan 2024, 15 January 2024
	regexp.MustCompile(`^\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4}$`),
	// DD MMM YYYY HH:MM: 15 Jan 2024 10:30
	regexp.MustCompile(`^\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4}\s+\d{1,2}:\d{2}(:\d{2})?$`),
	// DD MMM YYYY HH:MM AM/PM: 15 Jan 2024 10:30 AM
	regexp.MustCompile(`^\d{1,2}\s+[A-Za-z]{3,9}\s+\d{4}\s+\d{1,2}:\d{2}(:\d{2})?\s+(AM|PM|am|pm)$`),
	// Unix timestamp (10 or 13 digits)
	regexp.MustCompile(`^\d{10}$|^\d{13}$`),
}

// IsDateTime determines if content is a date or time.
func IsDateTime(content []byte) bool {
	text, ok := validateAndTrim(content)
	if !ok {
		return false
	}

	for _, re := range dateTimeRegexes {
		if re.MatchString(text) {
			return true
		}
	}

	return false
}
