// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package union

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kmpm/unionfees/internal"
	"github.com/kmpm/unionfees/public/spec"
	"github.com/shopspring/decimal"
)

var specEx1 spec.S1Spec = spec.S1Spec{
	CompanyNum:      5562344639,
	CompanyName:     "MAGNETBANDS REDOVISNING",
	Period:          4,
	LocNum:          1,
	Year:            12,
	TransactionDate: time.Date(2012, 04, 25, 0, 0, 0, 0, time.Local),
}

var dataEx1 []spec.S2Spec = []spec.S2Spec{
	{LocNum: 1, PersonNum: 1234567890, Name: "KARLSSON ALLAN", Amount: decimal.New(57035, -2), ControlAmount: decimal.New(0, 1), PayCode: spec.PayCodeAmountPayed},
	{LocNum: 1, PersonNum: 987654321, Name: "Johansson Evert", Amount: decimal.RequireFromString("640.00"), ControlAmount: decimal.New(0, 1), PayCode: spec.PayCodeAmountPayed},
	{LocNum: 1, PersonNum: 1122334455, Name: "Marklund Petrånella", Amount: decimal.New(0, 1), ControlAmount: decimal.New(0, 1), PayCode: spec.PayCodeEndEmployment},
}

var testLocations spec.Locations

func init() {
	testLocations = internal.BuildLocations(
		internal.CompanyArgs{
			CompanyNum:      5562344639,
			CompanyName:     "MAGNETBANDS REDOVISNING",
			Period:          4,
			Year:            12,
			TransactionDate: time.Date(2012, 04, 25, 0, 0, 0, 0, time.Local),
		},
		dataEx1,
	)
}

func np(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
func npByte(s byte) string {
	return np(string(s))
}

func TestWriteTable(t *testing.T) {
	type args struct {
		filename string
		locs     spec.Locations
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		matchAgainst string
	}{
		// TODO: Add test cases.
		{"first", args{"../../testdata/Metall-test.txt", testLocations}, false, "../../testdata/sample.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Create(tt.args.filename)
			if err != nil {
				t.Fatalf("error creating file %s: %v", tt.args.filename, err)
			}
			defer f.Close()

			if err := WriteTable(f, tt.args.locs, CodeIFMetall); (err != nil) != tt.wantErr {
				t.Fatalf("WriteTable() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := os.ReadFile(tt.args.filename)
			if err != nil {
				t.Fatalf("could not read 'got' file, error = %v", err)
			}
			want, err := os.ReadFile(tt.matchAgainst)
			if err != nil {
				t.Fatalf("could not read 'want' file, error = %v", err)
			}
			lg := len(got)

			for i, wb := range want {
				if i >= lg {
					t.Fatalf("got %d bytes, want %d bytes", lg, len(want))
				}
				gb := got[i]
				if gb != wb {
					t.Fatalf("byte %d does not match. got %d(%s), want %d(%s)", i, gb, npByte(gb), wb, npByte(wb))
				}
			}
		})
	}
}
