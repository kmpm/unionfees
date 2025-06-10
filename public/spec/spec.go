// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package spec

import (
	"time"

	"github.com/shopspring/decimal"
)

type PayCode int

const (
	PayCodeAmountPayed       PayCode = 1
	PayCodeTimeOff           PayCode = 3
	PayCodeOther             PayCode = 8
	PayCodeEndEmployment     PayCode = 19
	PayCodeMissingPermission PayCode = 33
)

type S1Spec struct {
	LocNum          int
	CompanyNum      int
	CompanyName     string
	Period          int
	Year            int
	TransactionDate time.Time
}

type S2Spec struct {
	LocNum        int
	PersonNum     int
	Name          string
	Amount        decimal.Decimal
	ControlAmount decimal.Decimal
	PayCode       PayCode
}

type S3Spec struct {
	LocNum           int
	CompanyNum       int
	CompanyName      string
	Records          int
	SumAmout         decimal.Decimal
	SumControlAmount decimal.Decimal
}

type Spec struct {
	S1 S1Spec
	S2 []S2Spec
	S3 S3Spec
}

type Locations map[int]Spec
