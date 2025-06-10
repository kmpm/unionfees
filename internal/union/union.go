// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package union

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/kmpm/unionfees/public/spec"
	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type UnionCode int

const (
	CodeIFMetall UnionCode = 38 // Union number for IF-Metall
	CodeGSUnion  UnionCode = 43 // Union number for GS-Unions
)

func (c UnionCode) String() string {
	switch c {
	case CodeIFMetall:
		return "IF-Metall"
	case CodeGSUnion:
		return "GS-facket"
	default:
		return "Unknown Union"
	}
}

type writer struct {
	buf *bufio.Writer
}

// padRight trims and pads string to n characters
func padRight(v string, n int) string {
	if len(v) > n {
		v = v[0:n]
	}
	format := fmt.Sprintf("%%-%ds", n)
	return fmt.Sprintf(format, v)
}

func padZero(v int, n int) string {
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, v)
}

func padDecimal(d decimal.Decimal, ni, nd int) string {
	v := strings.Split(d.StringFixed(2), ".")
	a, err := strconv.Atoi(v[0])
	if err != nil {
		log.Fatal(err)
	}
	b, err := strconv.Atoi(v[1])
	if err != nil {
		log.Fatal(err)
	}

	return padZero(a, ni) + padZero(b, nd)
}

func padDate(v time.Time) string {
	return fmt.Sprintf("%02d%02d%02d",
		v.Year()-2000,
		v.Month(),
		v.Day(),
	)
}

func (w *writer) writeStr(s string) error {
	_, err := w.buf.WriteString(s)
	if err != nil {
		log.Fatalf("error writing %s: %v", s, err)
	}
	return err
}

func (w *writer) writeStrR(s string, n int) error {
	return w.writeStr(padRight(strings.ToUpper(s), n))
}

func (w *writer) writeInt(v, n int) error {
	return w.writeStr(padZero(v, n))
}

// writeAmount with zero padded i characters for integer part
// and d characters for decimal part
func (w *writer) writeAmount(v decimal.Decimal, i, d int) error {
	return w.writeStr(padDecimal(v, i, d))
}

func (w *writer) writeS1(data spec.S1Spec, unionNo int) {
	w.writeStr("S1")
	w.writeInt(unionNo, 2)
	w.writeInt(data.LocNum, 4)
	w.writeInt(data.CompanyNum, 10)
	w.writeStrR(data.CompanyName, 24)
	w.writeStr("0")
	w.writeInt(data.Period, 2)
	w.writeInt(data.Year, 2)
	w.writeStr(padDate(data.TransactionDate))
	w.writeStr("0000000000000\r\n")
}

func (w *writer) writeS2(data spec.S2Spec, unionNo int) {
	w.writeStr("S2")
	w.writeInt(unionNo, 2)
	w.writeInt(data.LocNum, 4)
	w.writeInt(data.PersonNum, 10)
	w.writeStr(padRight(strings.ToUpper(data.Name), 24))
	w.writeAmount(data.Amount, 4, 2)
	w.writeAmount(data.ControlAmount, 4, 2)

	w.writeInt(int(data.PayCode), 2)
	w.writeStr("0000000000\r\n")
	// fmt.Printf("%s = %s\n", data.Name, data.Amount.StringFixed(2))
}

func (w *writer) writeS3(data spec.S3Spec, unionNo int) {
	w.writeStr("S3")
	w.writeInt(unionNo, 2)
	w.writeInt(data.LocNum, 4)
	w.writeInt(data.CompanyNum, 10)
	w.writeStr(padRight(strings.ToUpper(data.CompanyName), 24))
	w.writeInt(data.Records, 6)
	w.writeAmount(data.SumAmout, 7, 2)
	w.writeStr("000000000\r\n")
}

func WriteTable(iw io.Writer, locs spec.Locations, unionNo UnionCode) error {

	// which encoding?
	enc := transform.NewWriter(iw, charmap.Windows1252.NewEncoder())
	w := writer{
		buf: bufio.NewWriter(enc),
	}
	for locnum := range locs {
		w.writeS1(locs[locnum].S1, int(unionNo))
		for _, s2 := range locs[locnum].S2 {
			w.writeS2(s2, int(unionNo))
		}
		w.writeS3(locs[locnum].S3, int(unionNo))
	}
	err := w.buf.Flush()

	return err
}

func Str2UnionCode(s string) (UnionCode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("union string is empty")
	}

	union, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("could not convert union %s to int: %w", s, err)
	}
	if union < 1 || union > 99 {
		return 0, fmt.Errorf("union number %d is out of range (1-99)", union)
	}
	return UnionCode(union), nil
}
