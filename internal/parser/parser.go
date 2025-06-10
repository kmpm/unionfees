// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package parser

import (
	"log/slog"

	"github.com/ledongthuc/pdf"
)

type Cols []pdf.Text

func ReadPdf(path string) (*Document, error) {
	doc := Document{}
	f, r, err := pdf.Open(path)
	// remember close file
	if err != nil {
		return &doc, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("error closing file", "filename", f.Name(), "error", err)
		} else {
			slog.Debug("file closed", "filename", f.Name())
		}
	}()
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		page := doc.CreatePage()

		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		var lastTextStyle pdf.Text
		content := p.Content()

		// rects := content.Rect
		// for _, r := range rects {
		// 	fmt.Printf("Rect: %+v\n", r)
		// }

		texts := content.Text
		var row Row
		rows := Rows{0: row}

		for _, text := range texts {
			if r, ok := rows[text.Y]; ok {
				// switch if not belonging to current
				if !row.BelongsToRow(text) {
					rows[lastTextStyle.Y] = row //store before switching
					row = r
				}
			} else {
				rows[lastTextStyle.Y] = row
				row = Row{}
				rows[text.Y] = row
			}
			lastTextStyle = row.Add(text)
		}
		page.Rows = rows
	}

	return &doc, nil
}
