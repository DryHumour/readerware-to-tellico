package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

var (
	bom = []byte{0xEF, 0xBB, 0xBF}
)

// csvCmd represents the csv command
var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Manipulate CSV data",
	Long:  `Manipulate CSV data in various ways.`,
}

// csvListCmd represents the CSV list command
var csvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the column names from a CSV file on stdin",
	Long:  `List the column names from a CSV file on stdin, one per line, as JSON strings.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runListCSV(cmd)
	},
}

// csvGetCmd represents the CSV get command
var csvGetCmd = &cobra.Command{
	Use:   "get <column-name>",
	Short: "Get a specific column from a CSV file on stdin",
	Long:  `Get a specific column from a CSV file on stdin, printing each row's value as a JSON string.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetCSV(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(csvCmd)
	csvCmd.AddCommand(csvListCmd)
	csvListCmd.Flags().BoolP("raw", "r", false, "Output raw UTF-8 column names without JSON encoding")
	csvCmd.AddCommand(csvGetCmd)
	csvGetCmd.Flags().BoolP("raw", "r", false, "Output raw UTF-8 column values without JSON encoding")
}

func runListCSV(cmd *cobra.Command) error {
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()

	reader := bufio.NewReader(in)

	peek, err := reader.Peek(3)
	if err == nil && bytes.Equal(peek, bom) {
		reader.Discard(3)
	}

	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}

	for _, col := range header {
		if raw {
			fmt.Fprintln(out, col)
		} else {
			jsonBytes, _ := json.Marshal(col)
			fmt.Fprintln(out, string(jsonBytes))
		}
	}

	return nil
}

func runGetCSV(cmd *cobra.Command, columnName string) error {
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()

	reader := bufio.NewReader(in)

	peek, err := reader.Peek(3)
	if err == nil && bytes.Equal(peek, bom) {
		reader.Discard(3)
	}

	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	colIndex := slices.Index(header, columnName)
	if colIndex == -1 {
		return fmt.Errorf("column %q not found in CSV header", columnName)
	}

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}

	for {
		row, err := csvReader.Read()
		if err != nil {
			break
		}
		if colIndex < len(row) {
			if raw {
				fmt.Fprintln(out, row[colIndex])
			} else {
				jsonBytes, _ := json.Marshal(row[colIndex])
				fmt.Fprintln(out, string(jsonBytes))
			}
		}
	}

	return nil
}
