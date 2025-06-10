// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kmpm/unionfees/internal"
	"github.com/kmpm/unionfees/internal/parser"
	"github.com/kmpm/unionfees/internal/union"
	"github.com/ledongthuc/pdf"
)

var (
	flagNum     string
	flagName    string
	flagPeriod  int
	flagYear    int
	flagDate    string
	flagPrint   bool
	flagVersion bool
)

var appVersion = "v0.0.0-dev"

func init() {
	flag.StringVar(&flagNum, "o", "", "organisationsnummer (default från pdf)")
	flag.StringVar(&flagName, "n", "", "företagsnamn (default från pdf)")
	flag.IntVar(&flagPeriod, "m", 0, "redovisningsperiod MM (default från datum)")
	flag.IntVar(&flagYear, "y", 0, "redovisningår ÅÅ (default från datum)")
	flag.StringVar(&flagDate, "d", "", "utbetalningsdatum ÅÅMMDD")
	flag.BoolVar(&flagPrint, "print", false, "Visa det tolkade dokumentet")
	flag.BoolVar(&flagVersion, "version", false, "Visa versionsnummer och avsluta")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n [flags] <filename.pdf>", os.Args[0])
	fmt.Fprint(os.Stderr, "\nFlags\n")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("Version %s", appVersion)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flagVersion {
		printVersion()
		os.Exit(0)
	}

	t, err := time.Parse("060102", flagDate)
	if err != nil {
		fmt.Printf("Felaktigt datum: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
	now := time.Now()
	if t.Year() < now.Year() {
		fmt.Printf("Felaktigt år i datum: %d < %d\n", t.Year(), now.Year())
		flag.Usage()
		os.Exit(1)
	}

	if !isFlagPassed("m") {
		flagPeriod = int(t.Month())
	}

	if flagPeriod < 1 || flagPeriod > 12 {
		fmt.Printf("Felaktig period/månad: %d\n", flagPeriod)
		flag.Usage()
		os.Exit(1)
	}

	if !isFlagPassed("y") {
		flagYear = t.Year() - 2000
	}

	if flagYear < 0 || flagYear > 99 || flagYear < (now.Year()-2000) {
		fmt.Printf("Felaktigt år: %d\n", flagYear)
		flag.Usage()
		os.Exit(1)
	}

	pdf.DebugOn = true
	filename := flag.Arg(0)
	if filename == "" {
		log.Fatal("filnamn för pdf måste anges")
	}
	doc, err := parser.ReadPdf(filename) // Read local pdf file
	if err != nil {
		log.Fatalf("fel vid läsning av pdf: %v", err)
	}

	if flagPrint {
		doc.Fprint(os.Stdout)
	}

	matrix := doc.Pages[0].Strings()
	if flagName == "" {
		flagName = matrix[0][0]
	}

	if len(flagName) < 3 {
		fmt.Printf("Namn måste vara längre än 3 tecken")
		os.Exit(1)
	}
	fmt.Printf("Företag: \t%s\n", flagName)

	if flagNum == "" {
		flagNum = matrix[1][0]
	}

	cn, err := internal.Str2Person(flagNum)
	if err != nil {
		fmt.Printf("Felaktigt orgnr: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Orgnr:   \t%s\n", flagNum)
	fmt.Printf("Utb. datum: \t%s\n", flagDate)
	fmt.Printf("År:     \t%d\n", flagYear)
	fmt.Printf("Månad:     \t%d\n", flagPeriod)

	for k, v := range doc.GetTables() {
		fmt.Println(k)

		listS2, err := internal.ConvertS2Data(1, v)
		if err != nil {
			log.Fatal("error generating S2 list", err)
		}

		locs := internal.BuildLocations(
			internal.CompanyArgs{
				CompanyNum:      cn,
				CompanyName:     flagName,
				Period:          flagPeriod,
				Year:            flagYear,
				TransactionDate: t,
			},
			listS2,
		)
		for _, l := range locs {
			fmt.Printf("Plats %d, Antal: %d, Summa: %s\n", l.S3.LocNum, len(l.S2), l.S3.SumAmout)
		}

		filename := fmt.Sprintf("%s-%02d%02d.txt", k, flagYear, flagPeriod)
		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf("error creating file %s: %v", filename, err)
		}
		defer f.Close()
		err = union.WriteTable(f, locs, union.CodeIFMetall)
		if err != nil {
			log.Fatalf("error writing %s: %v", k, err)
		}
		fmt.Printf("\nFilen '%s' är skapad\n", filename)
	}
}
