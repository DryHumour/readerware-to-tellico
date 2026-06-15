package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DryHumour/readerware-to-tellico/internal/convert"
	"github.com/spf13/cobra"
)

// templatesCmd represents the templates parent command
var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"template", "t"},
	Short:   "Manage and inspect default templates",
	Long: `Manages and inspects the built-in conversion templates.

You can list the names of all default templates, or export them to a local
directory to use as a starting point for your own custom overrides.`,
}

// listCmd represents the templates list sub-command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all default templates",
	Long:    `Lists the filenames of all built-in Go templates embedded in the tool.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runListTemplates(cmd)
	},
}

// exportCmd represents the templates export sub-command
var exportCmd = &cobra.Command{
	Use:     "export [template-name]",
	Aliases: []string{"extract", "ext", "export", "exp", "dump", "e", "x", "d"},
	Short:   "Export default templates to a directory or stdout",
	Long: `Exports built-in templates to a local folder or prints a specific template to stdout.

If a specific template name is provided (e.g., "books.config.gotmpl"):
  - It prints the content of that template to stdout.
  - If --output-dir is specified, it saves that single template to the folder instead.

If no template name is provided:
  - It exports all built-in templates to the directory specified by --output-dir (required).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportTemplates(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(listCmd)
	templatesCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("output-dir", "o", "", "Directory to export template files into")
	exportCmd.MarkFlagDirname("output-dir")
}

func runListTemplates(*cobra.Command) error {
	entries, err := fs.ReadDir(convert.TemplatesFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to read embedded templates: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gotmpl") {
			fmt.Println(entry.Name())
		}
	}
	return nil
}

func runExportTemplates(cmd *cobra.Command, args []string) error {
	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return err
	}
	outputDir = strings.TrimSpace(outputDir)
	if outputDir != "" {
		outputDir = filepath.Clean(outputDir)
	} else if cmd.Flags().Lookup("output-dir").Changed {
		return fmt.Errorf("the --output-dir flag must be non-empty if provided")
	}

	// Case 1: Export a specific template
	if len(args) == 1 {
		name := args[0]
		// Clean the path to prevent directory traversal
		name = filepath.Clean(name)
		if strings.Contains(name, "/") || strings.Contains(name, "\\") || name == ".." {
			return fmt.Errorf("invalid template name: %s", name)
		}

		filePath := "templates/" + name
		content, err := convert.TemplatesFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read embedded template %q: %w", name, err)
		}

		if outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
			}
			targetPath := filepath.Join(outputDir, name)
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write template to %q: %w", targetPath, err)
			}
			cmd.Printf("Exported %s to %s\n", name, targetPath)
		} else {
			fmt.Println(string(content))
		}
		return nil
	}

	// Case 2: Export all templates
	if outputDir == "" {
		return fmt.Errorf("the --output-dir flag is required when exporting all templates")
	}

	entries, err := fs.ReadDir(convert.TemplatesFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to read embedded templates: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gotmpl") {
			name := entry.Name()
			content, err := convert.TemplatesFS.ReadFile("templates/" + name)
			if err != nil {
				return fmt.Errorf("failed to read embedded template %q: %w", name, err)
			}
			targetPath := filepath.Join(outputDir, name)
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write template to %q: %w", targetPath, err)
			}
			count++
		}
	}
	cmd.Printf("Successfully exported %d templates to %s\n", count, outputDir)
	return nil
}
