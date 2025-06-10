// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func ensureDir(filename string) error {
	path := filepath.Dir(filename)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func kvFromStrings(data [][]string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(data); i++ {
		row := data[i]
		for j := 0; j < len(row); j++ {
			cell := row[j]
			if strings.Contains(cell, ":") {
				// split on first colon
				kv := strings.SplitN(cell, ":", 2)
				if len(kv) == 2 {
					if len(kv[1]) > 0 {
						m[kv[0]] = strings.Trim(kv[1], " \t\r\n")
					} else if j < len(row) && !strings.Contains(":", row[j+1]) {
						m[kv[0]] = strings.Trim(row[j+1], " \t\r\n")
					}
				}

			}
		}
	}
	return m
}
