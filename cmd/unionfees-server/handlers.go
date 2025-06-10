// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kmpm/unionfees/internal"
	"github.com/kmpm/unionfees/internal/parser"
	"github.com/kmpm/unionfees/internal/union"
)

func parseToMultiZipHandler(c *gin.Context) {
	formDate := c.PostForm("period")
	t, err := time.Parse("2006-01-02", formDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info(file.Filename)
	ext := filepath.Ext(file.Filename)
	err = ensureDir(inboundFolder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inb, err := os.CreateTemp(inboundFolder, fmt.Sprintf("file-*%s", ext))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func() {
		// make sure inbound file is closed and removed
		inb.Close()
		err := os.Remove(inb.Name())
		if err != nil {
			slog.Error("error removing file", "filename", inb.Name(), "error", err)
		} else {
			slog.Debug("file removed", "filename", inb.Name())
		}
	}()

	err = c.SaveUploadedFile(file, inb.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	doc, err := parser.ReadPdf(inb.Name()) // Read local pdf file
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	matrix := doc.Pages[0].Strings()
	kv := kvFromStrings(matrix)

	companyName := kv["Namn"]
	vatID := kv["Organisationsnr"]
	// periodStr := kv["Period"]

	cn, err := internal.Str2Person(vatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Debug("parsed document",
		"companyName", companyName,
		"vatID", vatID)

	now := time.Now()

	flagPeriod := int(t.Month())
	flagYear := t.Year() - 2000
	if flagYear < 0 || flagYear > 99 || flagYear < (now.Year()-2000) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Felaktigt år: %d", flagYear)})
		return
	}

	zbuff := new(bytes.Buffer)
	zf, err := NewZipFile(zbuff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for k, v := range doc.GetTables() {
		fmt.Println(k)

		listS2, err := internal.ConvertS2Data(1, v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		locs := internal.BuildLocations(
			internal.CompanyArgs{
				CompanyNum:      cn,
				CompanyName:     companyName,
				Period:          flagPeriod,
				Year:            flagYear,
				TransactionDate: t,
			},
			listS2,
		)
		for _, l := range locs {
			// fmt.Printf("Plats %d, Antal: %d, Summa: %s\n", l.S3.LocNum, len(l.S2), l.S3.SumAmout)
			slog.Info("Plats", "Nr", l.S3.LocNum, "Antal", len(l.S2), "Summa", l.S3.SumAmout)
		}

		filename := fmt.Sprintf("%s-%02d%02d.txt", k, flagYear, flagPeriod)

		buff := new(bytes.Buffer)

		err = union.WriteTable(buff, locs, union.CodeIFMetall)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		slog.Debug("file created", "filename", filename)
		_, err = zf.AddFile(filename, bufio.NewReader(buff))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	zf.Close()
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s-%02d%02d.zip", companyName, flagYear, flagPeriod),
	}
	c.DataFromReader(http.StatusOK, int64(zbuff.Len()), "application/zip", zbuff, extraHeaders)
}

func parseToFirstTxtHandler(c *gin.Context) {
	formUnionNo := c.PostForm("union")
	if formUnionNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unioNo is required"})
		return
	}
	unionNo, err := union.Str2UnionCode(formUnionNo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	formDate := c.PostForm("period")
	t, err := time.Parse("2006-01-02", formDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info(file.Filename)
	ext := filepath.Ext(file.Filename)
	err = ensureDir(inboundFolder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inb, err := os.CreateTemp(inboundFolder, fmt.Sprintf("file-*%s", ext))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func() {
		// make sure inbound file is closed and removed
		inb.Close()
		err := os.Remove(inb.Name())
		if err != nil {
			slog.Error("error removing file", "filename", inb.Name(), "error", err)
		} else {
			slog.Info("file removed", "filename", inb.Name())
		}
	}()

	err = c.SaveUploadedFile(file, inb.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	doc, err := parser.ReadPdf(inb.Name()) // Read local pdf file
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	matrix := doc.Pages[0].Strings()
	kv := kvFromStrings(matrix)

	companyName := kv["Namn"]
	vatID := kv["Organisationsnr"]
	// periodStr := kv["Period"]

	cn, err := internal.Str2Person(vatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("parsed document",
		"companyName", companyName,
		"vatID", vatID)

	now := time.Now()
	flagPeriod := int(t.Month())
	flagYear := t.Year() - 2000
	if flagYear < 0 || flagYear > 99 || flagYear < (now.Year()-2000) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Felaktigt år: %d", flagYear)})
		return
	}

	for k, v := range doc.GetTables() {

		if !strings.Contains(k, "Metall") {
			continue
		}

		listS2, err := internal.ConvertS2Data(1, v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		locs := internal.BuildLocations(
			internal.CompanyArgs{
				CompanyNum:      cn,
				CompanyName:     companyName,
				Period:          flagPeriod,
				Year:            flagYear,
				TransactionDate: t,
			},
			listS2,
		)
		for _, l := range locs {
			slog.Info("Plats", "Nr", l.S3.LocNum, "Antal", len(l.S2), "Summa", l.S3.SumAmout)
		}

		filename := fmt.Sprintf("%s-%02d%02d.txt", k, flagYear, flagPeriod)

		buff := new(bytes.Buffer)
		err = union.WriteTable(buff, locs, unionNo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf("attachment; filename=%s", filename),
		}
		c.DataFromReader(http.StatusOK, int64(buff.Len()), "text/plain", buff, extraHeaders)
		return
	}

}
