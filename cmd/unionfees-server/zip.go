// SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
//
// SPDX-License-Identifier: MIT
package main

import (
	"archive/zip"
	"io"
)

type ZipFile struct {
	z *zip.Writer
}

func NewZipFile(w io.Writer) (*ZipFile, error) {

	return &ZipFile{
		z: zip.NewWriter(w),
	}, nil
}

func (zf *ZipFile) AddFile(name string, r io.Reader) (written int64, err error) {
	f, err := zf.z.Create(name)
	if err != nil {
		return 0, err
	}

	written, err = io.Copy(f, r)
	return written, err
}

func (zf *ZipFile) Close() error {
	return zf.z.Close()
}
