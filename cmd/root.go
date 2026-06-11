package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "readerware-to-tellico",
	Short: "Readerware to Tellico Rescue Tool",
	Long: `Readerware to Tellico Rescue Tool 
	
A command-line application designed to help users migrate their collections
from Readerware to Tellico.  It reads a Readerware database, extracts
information, and generates a TC file that can be loaded into Tellico, ensuring
a smooth transition between the two applications.`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		v := viper.GetViper()
		if err := bindFlags(cmd, v); err != nil {
			return err
		}
		initLogging(cmd, v)
		return nil
	},
	Version: "1.0.0",
}

var envKeyReplacer = strings.NewReplacer("-", "_", ".", "_")

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// ExecuteContext adds all child commands to the root command and sets flags appropriately.
// This is called by main.main() with a context that can be cancelled by signals.
// It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) {
	rootCmd.SetContext(ctx)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rw2tc.yaml)")
	rootCmd.MarkPersistentFlagFilename("config", "yaml")
	rootCmd.PersistentFlags().String("log-level", "warn", "Set the log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	v := viper.GetViper()

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// 1. Check home directory.
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(home)
		} else if !errors.Is(err, fs.ErrNotExist) {
			cobra.CheckErr(err)
		}

		// 2. Check XDG/Windows AppData (The cleaner location)
		if confDir, err := os.UserConfigDir(); err == nil {
			// This maps to %AppData% on Windows and ~/.config on Linux
			v.AddConfigPath(filepath.Join(confDir, "rw2tc"))
		} else if !errors.Is(err, fs.ErrNotExist) {
			cobra.CheckErr(err)
		}

		// Search config with name ".rw2tc.yaml".
		v.SetConfigType("yaml")
		v.SetConfigName(".rw2tc")
	}

	v.SetEnvPrefix("RW2TC")
	v.SetEnvKeyReplacer(envKeyReplacer)
	v.AutomaticEnv() // read in environment variables that match

	for _, cmd := range rootCmd.Commands() {
		// ensure each sub-command has a default empty map in case the config file
		// doesn't contain a section for that command.
		v.SetDefault(cmd.Name(), map[string]any{})
	}

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed())
	}
}

// Bind each cobra flag to its associated viper configuration (config file, environment variable, etc.).
// (I really wish viper.Sub() were sane....)
func bindFlags(cmd *cobra.Command, v *viper.Viper) (err error) {
	path := ""
	if cmd.Parent() != nil {
		path = cmd.Name() + "."
	}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		name := path + f.Name
		// Bind the viper flag to the cobra flag.
		if err = v.BindPFlag(name, f); err != nil {
			err = fmt.Errorf("failed to bind flag %s: %w", name, err)
			return
		}
		// Apply viper value to the cobra flag if the user didn't set it via the CLI.
		if !f.Changed && v.IsSet(name) {
			val := v.Get(name)
			if err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				err = fmt.Errorf("failed to set flag %s: %w", f.Name, err)
				return
			}
		}
	})
	return err
}

func initLogging(cmd *cobra.Command, v *viper.Viper) {
	ns := cmd.Name()
	level := parseLogLevel(v.GetString(ns + ".log-level"))
	if level > slog.LevelInfo && v.GetBool(ns+".verbose") {
		level = slog.LevelInfo
	}
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.RFC3339Nano,
		NoColor:    false,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
