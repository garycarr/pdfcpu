/*
Copyright 2020 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/garycarr/pdfcpu/pkg/api"
	"github.com/garycarr/pdfcpu/pkg/cli"
	"github.com/garycarr/pdfcpu/pkg/log"
	"github.com/garycarr/pdfcpu/pkg/pdfcpu/model"
	"github.com/garycarr/pdfcpu/pkg/pdfcpu/types"
)

var inDir, outDir, resDir, fontDir, samplesDir string
var conf *model.Configuration

func isTrueType(filename string) bool {
	s := strings.ToLower(filename)
	return strings.HasSuffix(s, ".ttf") || strings.HasSuffix(s, ".ttc")
}

func userFonts(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	ff := []string(nil)
	for _, f := range files {
		if isTrueType(f.Name()) {
			fn := filepath.Join(dir, f.Name())
			ff = append(ff, fn)
		}
	}
	return ff, nil
}

func TestMain(m *testing.M) {
	inDir = filepath.Join("..", "..", "testdata")
	fontDir = filepath.Join(inDir, "fonts")
	resDir = filepath.Join(inDir, "resources")
	samplesDir = filepath.Join("..", "..", "samples")

	conf = api.LoadConfiguration()

	// Install test user fonts from pkg/testdata/fonts.
	fonts, err := userFonts(filepath.Join(inDir, "fonts"))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	if err := api.InstallFonts(fonts); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	if outDir, err = os.MkdirTemp("", "pdfcpu_cli_tests"); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	//fmt.Printf("outDir = %s\n", outDir)

	exitCode := m.Run()

	if err = os.RemoveAll(outDir); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func copyFile(t *testing.T, srcFileName, destFileName string) error {
	t.Helper()
	from, err := os.Open(srcFileName)
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.Create(destFileName)
	if err != nil {
		return err
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	return err
}

func imageFileNames(t *testing.T, dir string) []string {
	t.Helper()
	fn, err := model.ImageFileNames(dir, types.MB)
	if err != nil {
		t.Fatal(err)
	}
	return fn
}

func isPDF(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".pdf")
}

func allPDFs(t *testing.T, dir string) []string {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("pdfFiles from %s: %v\n", dir, err)
	}
	ff := []string(nil)
	for _, f := range files {
		if isPDF(f.Name()) {
			ff = append(ff, f.Name())
		}
	}
	return ff
}

func validateFile(t *testing.T, fileName string, conf *model.Configuration) error {
	t.Helper()
	_, err := cli.Process(cli.ValidateCommand([]string{fileName}, conf))
	return err
}

func TestValidate(t *testing.T) {
	msg := "TestValidateCommand"
	for _, f := range allPDFs(t, inDir) {
		inFile := filepath.Join(inDir, f)
		if err := validateFile(t, inFile, conf); err != nil {
			t.Fatalf("%s: %s: %v\n", msg, inFile, err)
		}
	}
}

func TestInfoCommand(t *testing.T) {
	msg := "TestInfoCommand"
	inFile := filepath.Join(inDir, "5116.DCT_Filter.pdf")

	cmd := cli.InfoCommand([]string{inFile}, nil, true, true, conf)
	if _, err := cli.Process(cmd); err != nil {
		t.Fatalf("%s: %v\n", msg, err)
	}
}

func TestUnknownCommand(t *testing.T) {
	msg := "TestUnknownCommand"
	inFile := filepath.Join(outDir, "go.pdf")

	cmd := &cli.Command{
		Mode:   99,
		InFile: &inFile,
		Conf:   conf}

	if _, err := cli.Process(cmd); err == nil {
		t.Fatalf("%s: %v\n", msg, err)
	}
}

// Enable this test for debugging of a specific file.
func XTestSomeCommand(t *testing.T) {
	msg := "TestSomeCommand"

	log.SetDefaultTraceLogger()
	//log.SetDefaultParseLogger()
	log.SetDefaultReadLogger()
	log.SetDefaultValidateLogger()
	log.SetDefaultOptimizeLogger()
	log.SetDefaultWriteLogger()
	//log.SetDefaultStatsLogger()

	inFile := filepath.Join(inDir, "test.pdf")

	cmd := cli.ValidateCommand([]string{inFile}, conf)

	if _, err := cli.Process(cmd); err != nil {
		t.Fatalf("%s %s: %v\n", msg, inFile, err)
	}
}