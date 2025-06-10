// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package parser

import (
	"github.com/ledongthuc/pdf"
)

type Row struct {
	Cols Cols
}

type Rows map[float64]Row

// Add to current row
func (r *Row) Add(t pdf.Text) pdf.Text {
	if r.Cols == nil {
		r.Cols = append(r.Cols, t)
		return t
	}

	last := &r.Cols[len(r.Cols)-1]

	if t.S == "\n" {
		return *last
	}
	if isSameColumn(&t, last, 5) {
		last.S += t.S
		last.W = t.X - last.X + t.W
		return *last
	}
	r.Cols = append(r.Cols, t)
	return t
}

// BelongsToRow compares to Y position of given row
func (r *Row) BelongsToRow(t pdf.Text) bool {
	return r.Cols[len(r.Cols)-1].Y == t.Y
}

func (r *Row) Strings() []string {
	s := make([]string, len(r.Cols))
	for j, t := range r.Cols {
		s[j] = t.S
	}
	return s
}

// isSameColumn compares X of pdf.Text and start of t2 must be in n number of charactes of t1.X + t1.W
func isSameColumn(t1, t2 *pdf.Text, n float64) bool {
	if t1.Font != t2.Font {
		return false
	}
	if t1.FontSize != t2.FontSize {
		return false
	}
	l := len(t2.S)
	w := t2.W / float64(l)

	acceptable := t2.X + t2.W + n*w
	return t1.X <= acceptable
}
