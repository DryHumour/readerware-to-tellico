package cmd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/DryHumour/readerware-to-tellico/internal/images/extract"
	"github.com/spf13/cobra"
)

// imagesCmd represents the images command group
var imagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image", "img", "i"},
	Short:   "Manage and extract Readerware images",
	Long: `Manages and extracts associated Readerware images.

You can automatically extract database image blobs directly from the internal
Readerware database, preparing them for conversion.`,
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:     "extract <db-path> <output-path>",
	Aliases: []string{"ext", "e", "x"},
	Short:   "Extract Readerware Images",
	Long: `Extracts binary image blobs directly from a Readerware HSQLDB database.

Readerware stores images internally inside its database files. This command 
reads those image blobs and writes them as files into the specified output 
directory, making them available to be mapped and referenced by subsequent 
conversion sub-commands (using --extracted-images-dir).

Arguments:
  db-path      Path to the Readerware database file or directory.
  output-path  Directory where the extracted images will be saved.

Example:
  readerware-to-tellico images extract /path/to/readerware.data /path/to/extracted_images/`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return Extract(cmd, args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(imagesCmd)
	imagesCmd.AddCommand(extractCmd)
	extractCmd.Flags().String("java-path", "", "Path to the Java executable to use")
}

func Extract(cmd *cobra.Command, dbPath string, outputPath string) error {
	ctx := cmd.Context()
	javaPath, err := cmd.Flags().GetString("java-path")
	if err != nil {
		return err
	}
	if err := extract.Images(ctx, dbPath, outputPath, javaPath); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("Readerware image extraction failed. The Java process exited with error: %w\n\n"+
				"Please verify that:\n"+
				"  1. The database path %q is correct and points to a valid Readerware HSQLDB database.\n"+
				"  2. The output directory %q exists and is writable.\n"+
				"  3. You have appropriate read/write permissions for these paths.", err, dbPath, outputPath)
		}

		// Otherwise, assume it's a Java execution/lookup failure and show the JRE help.
		return fmt.Errorf("%w\n\nReaderware image extraction requires a Java Runtime Environment (JRE).\n"+
			"Please perform one of the following actions:\n"+
			"  1. Install a JRE and make sure it is on your system PATH.\n"+
			"  2. If Readerware 4 is installed, ensure it is in its default directory (e.g. C:\\Program Files\\Readerware 4).\n"+
			"  3. Set the RW2TC_EXTRACT_JAVA_PATH environment variable to the path of your java executable.\n"+
			"  4. Provide the --java-path flag to this command with the absolute path to your java executable.", err)
	}
	return nil
}
