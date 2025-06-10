// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package parser

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Page struct {
	Rows Rows
}

func (p *Page) Strings() [][]string {

	keys := make([]float64, 0, len(p.Rows))
	for k := range p.Rows {
		keys = append(keys, k)
	}
	data := make([][]string, len(keys))

	sort.Sort(sort.Reverse(sort.Float64Slice(keys)))
	for i, k := range keys {
		row := p.Rows[k]
		data[i] = row.Strings()
		// fmt.Printf("[%d] %s\n", i, strings.Join(row.Strings(), ";"))
	}
	return data
}

func (p *Page) Fprint(w io.Writer) {
	for i, elem := range p.Strings() {
		fmt.Fprintf(w, " %02d: %s\n", i, strings.Join(elem, "; "))
	}
}
