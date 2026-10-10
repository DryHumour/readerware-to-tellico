package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/DryHumour/readerware-to-tellico/internal/images/extract"
)

// imagesCmd represents the images command group.
var imagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image", "img", "i"},
	Short:   "Manage and extract Readerware images",
	Long: `Manages and extracts associated Readerware images.

You can automatically extract database image blobs directly from the internal
Readerware database, preparing them for conversion.`,
}

// extractCmd represents the extract command.
var extractCmd = &cobra.Command{
	Use:     "extract <db-path> <output-path>",
	Aliases: []string{"ext", "export", "exp", "dump", "e", "x", "d"},
	Short:   "Extract Readerware Images",
	Long: `Extracts binary image blobs directly from a Readerware HSQLDB database.

Readerware stores images internally inside its database files. This command 
reads those image blobs and writes them as files into the specified output 
directory, making them available to be mapped and referenced by subsequent 
conversion sub-commands (using --extracted-images-dir).

Only the embedded HSQLDB database format is supported; Client/Server
deployments using an external database server cannot be read.

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

// Extract runs the images extract command, pulling images out of a Readerware
// database directory into outputPath using the ImageDumper tool.
func Extract(cmd *cobra.Command, dbPath, outputPath string) error {
	ctx := cmd.Context()
	javaPath, err := cmd.Flags().GetString("java-path")
	if err != nil {
		return err
	}
	if err := extract.Images(ctx, dbPath, outputPath, javaPath); err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
			return fmt.Errorf("%w\n\nPlease verify that the database path %q points to a valid Readerware HSQLDB database and that the output directory %q is writable", err, dbPath, outputPath)
		}

		if errors.Is(err, extract.ErrJavaNotFound) {
			var bundledHint string
			if runtime.GOOS == "windows" {
				bundledHint = " If Readerware 4 is installed in its default directory, its bundled JRE will be found automatically."
			}
			return fmt.Errorf("%w\n\nReaderware image extraction requires a Java Runtime Environment (JRE).%s\n"+
				"Install a JRE (or JDK), set JAVA_HOME, or provide the path to your java executable via --java-path or the RW2TC_IMAGES_EXTRACT_JAVA_PATH environment variable", err, bundledHint)
		}

		return err
	}
	return nil
}
