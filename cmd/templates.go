package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
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
	Aliases: []string{"extract", "ext", "exp", "dump", "e", "x", "d"},
	Short:   "Export default templates to a directory or stdout",
	Long: `Exports built-in templates to a local folder or prints a specific template to stdout.

If a specific template name is provided (e.g., "books.config.gotmpl"):
  - It prints the content of that template to stdout.
  - If --output-dir is specified, it saves that single template to the folder instead.

If no template name is provided:
  - It exports all built-in templates to the directory specified by --output-dir (required).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runExportAllTemplates(cmd)
		}
		return runExportTemplate(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(templatesCmd)
	templatesCmd.AddCommand(listCmd)
	templatesCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("output-dir", "o", "", "Directory to export template files into")
	exportCmd.MarkFlagDirname("output-dir")
}

func runListTemplates(cmd *cobra.Command) error {
	entries, err := fs.ReadDir(convert.TemplatesFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to read embedded templates: %w", err)
	}
	out := cmd.OutOrStdout()
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gotmpl") {
			fmt.Fprintln(out, entry.Name())
		}
	}
	return nil
}

func runExportTemplate(cmd *cobra.Command, name string) error {
	ctx := cmd.Context()
	logger := slog.Default()

	// Clean the path to forestall directory traversal
	name = strings.TrimSpace(path.Clean(filepath.ToSlash(name)))
	if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
		return fmt.Errorf("invalid template name: %q", name)
	}

	content, err := convert.TemplatesFS.ReadFile(path.Join("templates", name))
	if err != nil {
		return fmt.Errorf("failed to read embedded template %q: %w", name, err)
	}

	outputDir, err := templateOutputDir(cmd)
	if err != nil {
		return err
	}

	if outputDir == nil {
		if _, err := cmd.OutOrStdout().Write(content); err != nil {
			return fmt.Errorf("failed to write template to stdout: %w", err)
		}
		return nil
	}

	targetPath, err := filepath.Localize(name)
	if err != nil {
		//nolint:forbidigo // embedded template names should never fail to localize
		panic(fmt.Errorf("failed to localize template name %q: %w", name, err))
	}

	if err := outputDir.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write template to %q: %w", targetPath, err)
	}

	logger.InfoContext(ctx, "exported template", "name", name, "path", filepath.Join(outputDir.Name(), targetPath))
	return nil
}

func runExportAllTemplates(cmd *cobra.Command) error {
	ctx := cmd.Context()
	logger := slog.Default()

	outputDir, err := templateOutputDir(cmd)
	if err != nil {
		return err
	}
	if outputDir == nil {
		return fmt.Errorf("the --output-dir flag is required when exporting all templates")
	}

	entries, err := fs.ReadDir(convert.TemplatesFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to read embedded templates: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if err := context.Cause(ctx); err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gotmpl") {
			continue
		}
		name := entry.Name()
		filePath := path.Join("templates", name)
		content, err := convert.TemplatesFS.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read embedded template %q: %w", filePath, err)
		}
		targetPath, err := filepath.Localize(name)
		if err != nil {
			//nolint:forbidigo // embedded template names should never fail to localize
			panic(fmt.Errorf("failed to localize template name %q: %w", name, err))
		}
		fullPath := filepath.Join(outputDir.Name(), targetPath)
		if err := outputDir.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write template to %q: %w", fullPath, err)
		}
		logger.DebugContext(ctx, "exported template", "name", name, "path", fullPath)
		count++
	}

	logger.InfoContext(ctx, "successfully exported templates", "count", count, "output-dir", outputDir.Name())
	return nil
}

func templateOutputDir(cmd *cobra.Command) (*os.Root, error) {
	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		return nil, err
	}
	outputDir = strings.TrimSpace(outputDir)
	if outputDir != "" {
		outputDir = filepath.Clean(outputDir)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
		}
		return os.OpenRoot(outputDir)
	} else if cmd.Flags().Lookup("output-dir").Changed {
		return nil, fmt.Errorf("the --output-dir flag must be non-empty if provided")
	}
	return nil, nil
}
