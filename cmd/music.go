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

// musicCmd represents the music command
var musicCmd = &cobra.Command{
	Use:     "music",
	Aliases: []string{"m"},
	Short:   "Convert Readerware Music to Tellico",
	Long: `Converts a Readerware Music CSV export to a Tellico (.tc) collection file.

This command parses the exported CSV data, normalizes and cleans metadata fields 
(such as artist names, composers, record labels, and track lists), structures roles 
and confidence levels, and packages the results along with album covers into a 
single, fully-compliant Tellico collection file.

Example:
  readerware-to-tellico music -i export.csv -o music.tc --first-images-dir /path/to/covers`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return Music(cmd)
	},
}

func init() {
	rootCmd.AddCommand(musicCmd)
	musicCmd.PersistentFlags().StringP("input-file", "i", "export.csv", "Path to the exported Readerware CSV file (default is export.csv)")
	musicCmd.MarkPersistentFlagFilename("input-file", "csv")
	musicCmd.PersistentFlags().StringP("output-file", "o", "music.tc", "Path to the output TC file (default is music.tc)")
	musicCmd.MarkPersistentFlagFilename("output-file", "tc")
	musicCmd.PersistentFlags().StringP("first-images-dir", "I", "", "Directory containing the Readerware first images export (optional)")
	musicCmd.MarkPersistentFlagDirname("first-images-dir")
	musicCmd.PersistentFlags().String("second-images-dir", "", "Directory containing the Readerware second images export (optional)")
	musicCmd.MarkPersistentFlagDirname("second-images-dir")
	musicCmd.PersistentFlags().String("third-images-dir", "", "Directory containing the Readerware third images export (optional)")
	musicCmd.MarkPersistentFlagDirname("third-images-dir")
	musicCmd.PersistentFlags().String("fourth-images-dir", "", "Directory containing the Readerware fourth images export (optional)")
	musicCmd.MarkPersistentFlagDirname("fourth-images-dir")
	musicCmd.PersistentFlags().StringP("first-large-images-dir", "L", "", "Directory containing the Readerware first large images export (optional)")
	musicCmd.MarkPersistentFlagDirname("first-large-images-dir")
	musicCmd.PersistentFlags().String("second-large-images-dir", "", "Directory containing the Readerware second large images export (optional)")
	musicCmd.MarkPersistentFlagDirname("second-large-images-dir")
	musicCmd.PersistentFlags().String("third-large-images-dir", "", "Directory containing the Readerware third large images export (optional)")
	musicCmd.MarkPersistentFlagDirname("third-large-images-dir")
	musicCmd.PersistentFlags().String("fourth-large-images-dir", "", "Directory containing the Readerware fourth large images export (optional)")
	musicCmd.MarkPersistentFlagDirname("fourth-large-images-dir")
	musicCmd.PersistentFlags().String("extracted-images-dir", "", "Directory containing the images extracted using the extract sub-command (optional)")
	musicCmd.MarkPersistentFlagDirname("extracted-images-dir")
	musicCmd.PersistentFlags().StringSlice("template-dirs", nil, "Directories containing user-provided Tellico templates (optional)")
	musicCmd.MarkPersistentFlagDirname("template-dirs")
	musicCmd.PersistentFlags().Int("concurrency", 16, "Number of parallel readers for image copying (default is 16)")
}

func Music(cmd *cobra.Command) error {
	ctx := cmd.Context()
	logger := slog.Default()
	v := viper.GetViper()

	var dto struct{ Music config.DTO }
	if err := v.Unmarshal(&dto); err != nil {
		return fmt.Errorf("failed to unmarshal viper config: %w", err)
	}
	if err := dto.Music.Validate(); err != nil {
		return fmt.Errorf("invalid configuration:\n%w", err)
	}
	cfg, err := config.New(dto.Music, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize resources: %w", err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			logger.ErrorContext(ctx, "failed to cleanly close resources", "kind", collection.KindMusic, "error", closeErr)
		}
	}()

	converter, err := convert.NewConverter(ctx, cfg, collection.NewMusicPolicy(ctx, cfg))
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
