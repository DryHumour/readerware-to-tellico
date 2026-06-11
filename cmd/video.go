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

// videoCmd represents the video command
var videoCmd = &cobra.Command{
	Use:     "video",
	Aliases: []string{"v"},
	Short:   "Convert Readerware Video to Tellico",
	Long: `Converts a Readerware Video CSV export to a Tellico (.tc) collection file.

This command parses the exported CSV data, normalizes and cleans metadata fields 
(such as cast, directors, writers, and producers), structures film/TV-specific roles, 
and packages the results along with video cover art into a single, fully-compliant 
Tellico collection file.

Example:
  readerware-to-tellico video -i export.csv -o video.tc --first-images-dir /path/to/posters`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return Video(cmd)
	},
}

func init() {
	rootCmd.AddCommand(videoCmd)
	videoCmd.PersistentFlags().StringP("input-file", "i", "export.csv", "Path to the exported Readerware CSV file (default is export.csv)")
	videoCmd.MarkPersistentFlagFilename("input-file", "csv")
	videoCmd.PersistentFlags().StringP("output-file", "o", "video.tc", "Path to the output TC file (default is video.tc)")
	videoCmd.MarkPersistentFlagFilename("output-file", "tc")
	videoCmd.PersistentFlags().StringP("first-images-dir", "I", "", "Directory containing the Readerware first images export (optional)")
	videoCmd.MarkPersistentFlagDirname("first-images-dir")
	videoCmd.PersistentFlags().String("second-images-dir", "", "Directory containing the Readerware second images export (optional)")
	videoCmd.MarkPersistentFlagDirname("second-images-dir")
	videoCmd.PersistentFlags().String("third-images-dir", "", "Directory containing the Readerware third images export (optional)")
	videoCmd.MarkPersistentFlagDirname("third-images-dir")
	videoCmd.PersistentFlags().String("fourth-images-dir", "", "Directory containing the Readerware fourth images export (optional)")
	videoCmd.MarkPersistentFlagDirname("fourth-images-dir")
	videoCmd.PersistentFlags().StringP("first-large-images-dir", "L", "", "Directory containing the Readerware first large images export (optional)")
	videoCmd.MarkPersistentFlagDirname("first-large-images-dir")
	videoCmd.PersistentFlags().String("second-large-images-dir", "", "Directory containing the Readerware second large images export (optional)")
	videoCmd.MarkPersistentFlagDirname("second-large-images-dir")
	videoCmd.PersistentFlags().String("third-large-images-dir", "", "Directory containing the Readerware third large images export (optional)")
	videoCmd.MarkPersistentFlagDirname("third-large-images-dir")
	videoCmd.PersistentFlags().String("fourth-large-images-dir", "", "Directory containing the Readerware fourth large images export (optional)")
	videoCmd.MarkPersistentFlagDirname("fourth-large-images-dir")
	videoCmd.PersistentFlags().String("extracted-images-dir", "", "Directory containing the images extracted using the extract sub-command (optional)")
	videoCmd.MarkPersistentFlagDirname("extracted-images-dir")
	videoCmd.PersistentFlags().StringSlice("template-dirs", nil, "Directories containing user-provided Tellico templates (optional)")
	videoCmd.MarkPersistentFlagDirname("template-dirs")
	videoCmd.PersistentFlags().Int("concurrency", 16, "Number of parallel readers for image copying (default is 16)")
}

func Video(cmd *cobra.Command) error {
	ctx := cmd.Context()
	logger := slog.Default()
	v := viper.GetViper()

	var dto struct{ Video config.DTO }
	if err := v.Unmarshal(&dto); err != nil {
		return fmt.Errorf("failed to unmarshal viper config: %w", err)
	}
	if err := dto.Video.Validate(); err != nil {
		return fmt.Errorf("invalid configuration:\n%w", err)
	}
	cfg, err := config.New(dto.Video, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize resources: %w", err)
	}
	defer func() {
		if closeErr := cfg.Close(); closeErr != nil {
			logger.ErrorContext(ctx, "failed to cleanly close resources", "kind", collection.KindVideo, "error", closeErr)
		}
	}()

	converter, err := convert.NewConverter(ctx, cfg, collection.NewVideoPolicy(ctx, cfg))
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
