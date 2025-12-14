package config

// Copyright (C) 2025 Maxim Kim (exynil)
// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const configFile = "config.conf"

type Config struct {
	Limit                  int
	PinnedMarker           string
	UnpinnedMarker         string
	BufferNames            [5]string // Names for buffers 1-5
	SeparatorLength        int       // Length of separator between regular and pinned entries
	MaxDedupeSearch        int       // Maximum number of recent entries to check for duplicates
	MaxItems               int       // Maximum number of items to store (0 = unlimited)
	MinStoreLength         int       // Minimum number of characters to store
	DBPath                 string    // Path to database (empty = use default)
	PreviewWidth           int       // Maximum number of characters to preview
	ShowImageIcons         bool      // Show image icons in rofi (default: true)
	MaskPasswords          int       // Password masking mode: 0 = no masking, 1 = partial, 2 = full (default: 0)
	PasswordMaskColor      string    // Color for masked password characters (default: "red")
	PasswordMaskChar       string    // Character used for masking passwords (default: "*")
	PasswordIgnorePatterns []string  // Regex patterns to exclude from password detection
}

// GetConfigPath returns the path to the config file.
// Uses $XDG_CONFIG_HOME/clipbox/config.conf or ~/.config/clipbox/config.conf
func GetConfigPath() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	configPath := filepath.Join(configHome, "clipbox", configFile)
	return configPath, nil
}

// LoadConfig reads the configuration file and returns Config with values.
// Returns default values if config file doesn't exist or on error.
func LoadConfig() (*Config, error) {
	config := &Config{
		Limit:                  500,
		PinnedMarker:           "",
		UnpinnedMarker:         "",
		BufferNames:            [5]string{"", "", "", "", ""},
		SeparatorLength:        66,
		MaxDedupeSearch:        100,
		MaxItems:               500,
		MinStoreLength:         0,
		DBPath:                 "",
		PreviewWidth:           65,
		ShowImageIcons:         false,
		MaskPasswords:          0,
		PasswordMaskColor:      "#DC2626",
		PasswordMaskChar:       "*",
		PasswordIgnorePatterns: []string{},
	}

	configPath, err := GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get config path: %v, using defaults\n", err)
		return config, nil
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create config dir: %v, using defaults\n", err)
		return config, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		fmt.Fprintf(os.Stderr, "Warning: failed to open config: %v, using defaults\n", err)
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "limit":
			if limit, err := strconv.Atoi(value); err == nil && limit > 0 {
				config.Limit = limit
			}
		case "pinned_marker":
			config.PinnedMarker = value
		case "unpinned_marker":
			config.UnpinnedMarker = value
		case "buffer_1_name":
			config.BufferNames[0] = value
		case "buffer_2_name":
			config.BufferNames[1] = value
		case "buffer_3_name":
			config.BufferNames[2] = value
		case "buffer_4_name":
			config.BufferNames[3] = value
		case "buffer_5_name":
			config.BufferNames[4] = value
		case "separator_length":
			if length, err := strconv.Atoi(value); err == nil && length > 0 {
				config.SeparatorLength = length
			}
		case "max_dedupe_search":
			if maxSearch, err := strconv.Atoi(value); err == nil && maxSearch > 0 {
				config.MaxDedupeSearch = maxSearch
			}
		case "max_items":
			if maxItems, err := strconv.Atoi(value); err == nil && maxItems >= 0 {
				config.MaxItems = maxItems
			}
		case "min_store_length":
			if minLength, err := strconv.Atoi(value); err == nil && minLength >= 0 {
				config.MinStoreLength = minLength
			}
		case "db_path":
			config.DBPath = os.ExpandEnv(value)
		case "preview_width":
			if width, err := strconv.Atoi(value); err == nil && width > 0 {
				config.PreviewWidth = width
			}
		case "show_image_icons":
			switch value {
			case "false", "0", "no":
				config.ShowImageIcons = false
			case "true", "1", "yes":
				config.ShowImageIcons = true
			}
		case "mask_passwords":
			if mode, err := strconv.Atoi(value); err == nil {
				if mode >= 0 && mode <= 2 {
					config.MaskPasswords = mode
				}
			}
		case "password_mask_color":
			if value != "" {
				config.PasswordMaskColor = value
			}
		case "password_mask_char":
			if value != "" {
				// Use first rune if multiple characters provided
				runes := []rune(value)
				if len(runes) > 0 {
					config.PasswordMaskChar = string(runes[0])
				}
			}
		case "password_ignore_pattern":
			if value != "" {
				config.PasswordIgnorePatterns = append(config.PasswordIgnorePatterns, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return config, fmt.Errorf("error reading config: %w", err)
	}

	return config, nil
}

// GetDBPath returns the path to the SQLite database file.
// Uses custom path from config if set, otherwise defaults to $XDG_CACHE_HOME/clipbox/clipbox.db
func GetDBPath() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.DBPath != "" {
		return cfg.DBPath, nil
	}

	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheHome = filepath.Join(homeDir, ".cache")
	}
	dbPath := filepath.Join(cacheHome, "clipbox", "clipbox.db")
	return dbPath, nil
}
