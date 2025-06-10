// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kmpm/unionfees/public/spec"
	"github.com/shopspring/decimal"
)

var reParenthesis = regexp.MustCompile(`\((.*?)\)`)

func Str2Person(v string) (int, error) {
	v = strings.ReplaceAll(v, "-", "")
	return strconv.Atoi(v)
}

func str2Amount(v string) (decimal.Decimal, error) {
	v = strings.ReplaceAll(v, "-", "")
	v = strings.ReplaceAll(v, ",", ".")
	return decimal.NewFromString(v)
}

func name2LastFirst(name string) string {
	parts := strings.Split(name, " ")
	l := len(parts)
	lastname := parts[l-1]
	firstname := strings.Join(parts[:l-1], " ")
	return fmt.Sprintf("%s %s", lastname, firstname)
}

func cleanName(name string) string {
	all := reParenthesis.FindAllString(name, -1)
	for _, ele := range all {
		name = strings.ReplaceAll(name, ele, "")
	}
	return name
}

func rowS2Data(data []string, spec *spec.S2Spec) error {
	name := cleanName(data[1])
	spec.Name = name2LastFirst(name)
	n, err := Str2Person(data[2])
	if err != nil {
		return err
	}
	spec.PersonNum = n
	f, err := str2Amount(data[3])
	if err != nil {
		return err
	}
	spec.Amount = f
	return nil
}

func ConvertS2Data(locNum int, data [][]string) ([]spec.S2Spec, error) {
	specs := []spec.S2Spec{}
	for _, x := range data {
		s := spec.S2Spec{
			LocNum:  locNum,
			PayCode: spec.PayCodeAmountPayed,
		}
		err := rowS2Data(x, &s)
		if err != nil {
			return specs, err
		}
		specs = append(specs, s)
	}
	return specs, nil
}

type CompanyArgs struct {
	CompanyNum      int
	CompanyName     string
	Period          int
	Year            int
	TransactionDate time.Time
}

func buildS1(args CompanyArgs, locnum int) spec.S1Spec {
	return spec.S1Spec{
		LocNum:          locnum,
		CompanyNum:      args.CompanyNum,
		CompanyName:     args.CompanyName,
		Period:          args.Period,
		Year:            args.Year,
		TransactionDate: args.TransactionDate,
	}
}

func BuildLocations(args CompanyArgs, s2s []spec.S2Spec) spec.Locations {
	locs := spec.Locations{}
	for _, s2 := range s2s {
		locnum := s2.LocNum
		loc, ok := locs[locnum]
		if !ok {
			loc = spec.Spec{
				S1: buildS1(args, locnum),
				S2: []spec.S2Spec{},
				S3: spec.S3Spec{
					CompanyNum:  args.CompanyNum,
					CompanyName: args.CompanyName,
					LocNum:      locnum,
				},
			}
		}
		loc.S2 = append(loc.S2, s2)
		loc.S3.Records += 1
		loc.S3.SumAmout = s2.Amount.Add(loc.S3.SumAmout)
		loc.S3.SumControlAmount = s2.ControlAmount.Add(loc.S3.SumControlAmount)
		locs[locnum] = loc
	}
	return locs
}
