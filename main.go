package main

// Copyright (C) 2025 Maxim Kim (exynil)
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

import (
	"fmt"
	"os"
	"strconv"

	"clipbox/database"
	"clipbox/maintenance"
	"clipbox/utils"
)

const version = "0.1.0"

func main() {
	// Handle --version and -v flags first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("clipbox %s\n", version)
			os.Exit(0)
		}
	}

	rofiRetv := os.Getenv("ROFI_RETV")

	// Handle rofi kb-custom keys
	if rofiRetv != "" {
		retv, err := strconv.Atoi(rofiRetv)
		if err == nil {
			switch retv {
			case 10: // kb-custom-1: toggle pin
				if len(os.Args) >= 2 && os.Args[1] != "" {
					id, err := utils.ExtractID(os.Args[1])
					if err == nil && id > 0 {
						if err := database.TogglePin(id); err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}
						if err := database.List(0); err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}
						return
					}
				}
				if err := database.List(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			case 11, 12, 13, 14, 15: // kb-custom-2 to kb-custom-6: switch buffer
				bufferID := retv - 10
				if err := database.SwitchBuffer(bufferID); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if err := database.List(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			case 16: // kb-custom-7: delete entry
				if len(os.Args) >= 2 && os.Args[1] != "" {
					id, err := utils.ExtractID(os.Args[1])
					if err == nil && id > 0 {
						if err := database.DeleteEntry(id); err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}
						if err := database.List(0); err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}
						return
					}
				}
				if err := database.List(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			case 17: // kb-custom-8: switch to previous buffer
				if err := database.SwitchToPreviousBuffer(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if err := database.List(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			case 18: // kb-custom-9: switch to next buffer
				if err := database.SwitchToNextBuffer(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if err := database.List(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return
			}
		}
	}

	var command string
	if len(os.Args) < 2 {
		command = "--list"
	} else {
		command = os.Args[1]
	}

	switch command {
	case "--store":
		if err := database.Store(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "--list":
		limit := 0
		if len(os.Args) > 2 {
			if _, err := fmt.Sscanf(os.Args[2], "%d", &limit); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid limit: %s\n", os.Args[2])
				os.Exit(1)
			}
		}
		if err := database.List(limit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "--rebuild-previews":
		if err := maintenance.RebuildAllPreviews(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "--vacuum":
		if err := maintenance.VacuumDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		// Rofi modi mode: handle selection
		if len(os.Args) >= 2 && command != "" {
			id, err := utils.ExtractID(command)
			if err == nil && id > 0 {
				content, err := database.GetContentByID(id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if err := utils.CopyToClipboard(content); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			}
		}
		if err := database.List(0); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
