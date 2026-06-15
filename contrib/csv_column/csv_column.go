package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

var bom = []byte{0xEF, 0xBB, 0xBF}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [column_name]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  If column_name is provided, prints values for that column\n")
		fmt.Fprintf(os.Stderr, "  If no column_name is provided, prints all column names\n")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	peek, err := reader.Peek(3)
	if err == nil && bytes.Equal(peek, bom) {
		reader.Discard(3)
	}

	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CSV header: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		for _, col := range header {
			jsonBytes, _ := json.Marshal(col)
			fmt.Println(string(jsonBytes))
		}
		return
	}

	columnName := os.Args[1]
	colIndex := slices.Index(header, columnName)
	if colIndex == -1 {
		fmt.Fprintf(os.Stderr, "Column '%s' not found in CSV header\n", columnName)
		fmt.Fprintf(os.Stderr, "Available columns: %v\n", header)
		os.Exit(1)
	}

	for {
		row, err := csvReader.Read()
		if err != nil {
			break
		}
		if colIndex < len(row) {
			jsonBytes, _ := json.Marshal(row[colIndex])
			fmt.Println(string(jsonBytes))
		}
	}
}
