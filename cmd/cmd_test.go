package cmd

import (
	"fmt"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/build"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/spf13/pflag"
)

func resetBuildParams() {
	build.BuildParams = &build.BuildArgs{}
}

func resetCommands() {
	// Reset all command flags to their default values
	rootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := flag.Value.Set(flag.DefValue)
		if err != nil {
			panic(fmt.Sprintf("Failed to reset flag %s: %v", flag.Name, err))
		}
		flag.Changed = false
	})
	for _, cmd := range rootCmd.Commands() {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			err := flag.Value.Set(flag.DefValue)
			if err != nil {
				panic(fmt.Sprintf("Failed to reset flag %s: %v", flag.Name, err))
			}
			flag.Changed = false
		})
	}
}

func TestBuildHelp(t *testing.T) {
	defer resetBuildParams()
	defer resetCommands()
	logger.InitLogger()
	rootCmd.SetArgs([]string{"build", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no error when running 'build --help', got: %v", err)
	}
}

func TestBuildMissingToc(t *testing.T) {
	defer resetBuildParams()
	defer resetCommands()
	logger.InitLogger()
	rootCmd.SetArgs([]string{"build", "-t", "."})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error when running 'build' without TOC file, got nil")
	}
}
