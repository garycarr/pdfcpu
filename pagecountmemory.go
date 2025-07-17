package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/garycarr/pdfcpu/pkg/api"
	"github.com/garycarr/pdfcpu/pkg/pdfcpu/model"
)

func measureMemory() uint64 {
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func testPageCountOptimized(filename string) {
	fmt.Printf("Testing OPTIMIZED PageCount (ReadContext + EnsurePageCount) on %s\n", filename)

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	conf := model.NewDefaultConfiguration()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	start := time.Now()
	pageCount, err := api.PageCount(f, conf)
	duration := time.Since(start)

	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc

	if err != nil {
		fmt.Printf("Error getting page count: %v\n", err)
		return
	}

	fmt.Printf("Page count: %d\n", pageCount)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Memory before: %d bytes (%.2f MB)\n", memBefore, float64(memBefore)/1024/1024)
	fmt.Printf("Memory after: %d bytes (%.2f MB)\n", memAfter, float64(memAfter)/1024/1024)
	if memAfter > memBefore {
		fmt.Printf("Memory used: %d bytes (%.2f MB)\n", memAfter-memBefore, float64(memAfter-memBefore)/1024/1024)
	} else {
		fmt.Printf("Memory used: minimal (GC occurred during operation)\n")
	}
	fmt.Println("---")
}

func testPageCountFull(filename string) {
	fmt.Printf("Testing ORIGINAL PageCount (ReadAndValidate) on %s\n", filename)

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	conf := model.NewDefaultConfiguration()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	start := time.Now()
	ctx, err := api.ReadAndValidate(f, conf)
	if err != nil {
		fmt.Printf("Error reading PDF: %v\n", err)
		return
	}
	pageCount := ctx.PageCount
	duration := time.Since(start)

	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc

	fmt.Printf("Page count: %d\n", pageCount)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Memory before: %d bytes (%.2f MB)\n", memBefore, float64(memBefore)/1024/1024)
	fmt.Printf("Memory after: %d bytes (%.2f MB)\n", memAfter, float64(memAfter)/1024/1024)
	if memAfter > memBefore {
		fmt.Printf("Memory used: %d bytes (%.2f MB)\n", memAfter-memBefore, float64(memAfter-memBefore)/1024/1024)
	} else {
		fmt.Printf("Memory used: minimal (GC occurred during operation)\n")
	}
	fmt.Println("---")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_pagecount_memory.go <pdf_file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	fmt.Println("=== MEMORY USAGE COMPARISON ===")
	testPageCountOptimized(filename)
	testPageCountFull(filename)
}
