// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package parser

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type Document struct {
	Pages       []*Page
	CompanyName string
	CompanyNum  int
	mu          sync.Mutex
}

func (d *Document) CreatePage() *Page {
	p := &Page{}
	d.AddPage(p)
	return p
}

func (d *Document) AddPage(p *Page) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Pages = append(d.Pages, p)
}

func (d *Document) Fprint(w io.Writer) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, p := range d.Pages {
		fmt.Fprintf(w, "--- Page %d ---\n", i)
		p.Fprint(w)
	}
}

func (d *Document) GetTables() map[string][][]string {
	tables := map[string][][]string{}
	var table [][]string
	var tablename string
	var start, end int
	for _, p := range d.Pages {
		data := p.Strings()
		for i, r := range data {
			if start == 0 || i < start {
				for _, s := range r {
					if strings.HasPrefix(s, "Fackförbund: ") {
						if tablename != "" {
							tables[tablename] = table
						}

						tablename = strings.Trim(strings.Split(s, ":")[1], " \t")
						table = [][]string{}
						start = i + 2
						end = 0
					}
				}
			}
			if (start > 0 && i >= start) && (end == 0 || i > end) {
				if len(r) == 4 {
					table = append(table, r)
				}
				end = i
			}
		}
	}
	if tablename != "" {
		tables[tablename] = table
	}
	return tables
}
