package cmd

import (
	"fmt"
	"log/slog"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/convert"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// booksCmd represents the books command
var booksCmd = &cobra.Command{
	Use:     "books",
	Aliases: []string{"b"},
	Short:   "Convert Readerware Books to Tellico",
	Long: `Converts a Readerware Books CSV export to a Tellico (.tc) collection file.

This command parses the exported CSV data, normalizes and cleans metadata fields 
(such as names, ISBNs, bindings, and conditions), structures roles (such as authors, 
editors, and illustrators), and packages the results along with any referenced 
images into a single, fully-compliant Tellico collection file.

Example:
  readerware-to-tellico books -i export.csv -o books.tc --first-images-dir /path/to/images`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return Books(cmd)
	},
}

func init() {
	rootCmd.AddCommand(booksCmd)
	booksCmd.PersistentFlags().StringP("input-file", "i", "export.csv", "Path to the exported Readerware CSV file (default is export.csv)")
	booksCmd.MarkPersistentFlagFilename("input-file", "csv")
	booksCmd.PersistentFlags().StringP("output-file", "o", "books.tc", "Path to the output TC file (default is books.tc)")
	booksCmd.MarkPersistentFlagFilename("output-file", "tc")
	booksCmd.PersistentFlags().StringP("first-images-dir", "I", "", "Directory containing the Readerware first images export (optional)")
	booksCmd.MarkPersistentFlagDirname("first-images-dir")
	booksCmd.PersistentFlags().String("second-images-dir", "", "Directory containing the Readerware second images export (optional)")
	booksCmd.MarkPersistentFlagDirname("second-images-dir")
	booksCmd.PersistentFlags().String("third-images-dir", "", "Directory containing the Readerware third images export (optional)")
	booksCmd.MarkPersistentFlagDirname("third-images-dir")
	booksCmd.PersistentFlags().String("fourth-images-dir", "", "Directory containing the Readerware fourth images export (optional)")
	booksCmd.MarkPersistentFlagDirname("fourth-images-dir")
	booksCmd.PersistentFlags().StringP("first-large-images-dir", "L", "", "Directory containing the Readerware first large images export (optional)")
	booksCmd.MarkPersistentFlagDirname("first-large-images-dir")
	booksCmd.PersistentFlags().String("second-large-images-dir", "", "Directory containing the Readerware second large images export (optional)")
	booksCmd.MarkPersistentFlagDirname("second-large-images-dir")
	booksCmd.PersistentFlags().String("third-large-images-dir", "", "Directory containing the Readerware third large images export (optional)")
	booksCmd.MarkPersistentFlagDirname("third-large-images-dir")
	booksCmd.PersistentFlags().String("fourth-large-images-dir", "", "Directory containing the Readerware fourth large images export (optional)")
	booksCmd.MarkPersistentFlagDirname("fourth-large-images-dir")
	booksCmd.PersistentFlags().String("extracted-images-dir", "", "Directory containing the images extracted using the extract sub-command (optional)")
	booksCmd.MarkPersistentFlagDirname("extracted-images-dir")
	booksCmd.PersistentFlags().StringSlice("template-dirs", nil, "Directories containing user-provided Tellico templates (optional)")
	booksCmd.MarkPersistentFlagDirname("template-dirs")
	booksCmd.PersistentFlags().Int("concurrency", 16, "Number of parallel readers for image copying (default is 16)")
}

func Books(cmd *cobra.Command) error {
	ctx := cmd.Context()
	logger := slog.Default()
	v := viper.GetViper()

	var dto struct{ Books config.DTO }
	if err := v.Unmarshal(&dto); err != nil {
		return fmt.Errorf("failed to unmarshal viper config: %w", err)
	}
	if err := dto.Books.Validate(); err != nil {
		return fmt.Errorf("invalid configuration:\n%w", err)
	}
	cfg, err := config.New(dto.Books, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize resources: %w", err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			logger.ErrorContext(ctx, "failed to cleanly close resources", "kind", collection.KindBooks, "error", closeErr)
		}
	}()

	converter, err := convert.NewConverter(ctx, cfg, collection.NewBooksPolicy(ctx, cfg))
	if err != nil {
		return fmt.Errorf("failed to create converter: %w", err)
	}
	var errs []error
	for report, err := range converter.Run(ctx) {
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}
		report.Log(ctx, logger)
		if report.Err != nil && report.Level >= slog.LevelError {
			errs = append(errs, report.Err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("conversion failed with %d errors", len(errs))
	}

	return nil
}
