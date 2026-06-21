/*
Copyright © 2025 Rob "McTalian" Anderson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/McTalian/wow-build-tools/internal/cmdargs"
	"github.com/McTalian/wow-build-tools/internal/config"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wow-build-tools",
	Short: "World of Warcraft addon build tools",
	Long: `This tool is used to build World of Warcraft addons both for local development
	and for release/distribution via CurseForge, WoWInterface, and Wago.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func configureLogger() {
	if cmdargs.RootParams.NoEmoji || cmdargs.RootParams.Boring {
		logger.DisableEmoji()
	}
	if cmdargs.RootParams.NoColor || cmdargs.RootParams.Boring {
		color.NoColor = true
	}

	if cmdargs.RootParams.LevelVerbose {
		logger.SetLogLevel(logger.VERBOSE)
	} else if cmdargs.RootParams.LevelDebug {
		logger.SetLogLevel(logger.DEBUG)
	} else {
		logger.SetLogLevel(logger.INFO)
	}
}

func loadConfig() error {
	viper.SetConfigName(".wbt")
	viper.SetConfigType("yaml")
	configDir, err := config.GetConfigDir()
	if err != nil {
		logger.Error("Failed to determine configuration directory: %v - Please report this issue on GitHub", err)
		return err
	}
	// viper.AddConfigPath(".")       // Look for config file in current directory first
	// Maybe support merging multiple config files? For now just the global one is good enough
	viper.AddConfigPath(configDir) // Then look in the user's home directory

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Verbose("No configuration file (.wbt.yaml) found at %s or current directory", configDir)
		} else {
			logger.Error("Failed to read configuration file: %v", err)
			return err
		}
	}

	return nil
}

func preRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err != nil {
			logger.Error("Error during pre-run setup: %v", err)
		}
	}()

	configureLogger()
	err = loadConfig()
	if err != nil {
		return
	}

	if !cmdargs.RootParams.NoSplash && !cmdargs.RootParams.Boring {
		fmt.Println(`
===============================================================================
 __    __     __    __        ___       _ _     _      _____            _     
/ / /\ \ \___/ / /\ \ \      / __\_   _(_) | __| |    /__   \___   ___ | |___ 
\ \/  \/ / _ \ \/  \/ /____ /__\// | | | | |/ _  |_____ / /\/ _ \ / _ \| / __|
 \  /\  / (_) \  /\  /_____/ \/  \ |_| | | | (_| |_____/ / | (_) | (_) | \__ \
  \/  \/ \___/ \/  \/      \_____/\__,_|_|_|\__,_|     \/   \___/ \___/|_|___/

              />
 (           //------------------------------------------------------(
(*)OXOXOXOXO(*>                  --------                             \
 (           \\--------------------------------------------------------)
              \>
===============================================================================`)
	}

	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wow-build-tools.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&cmdargs.RootParams.LevelVerbose, "verbose", "V", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&cmdargs.RootParams.LevelDebug, "debug", "v", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVar(&cmdargs.RootParams.NoEmoji, "no-emoji", false, "Disable emoji output")
	rootCmd.PersistentFlags().BoolVar(&cmdargs.RootParams.NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&cmdargs.RootParams.NoSplash, "no-splash", false, "Disable splash screen on startup")
	rootCmd.PersistentFlags().BoolVar(&cmdargs.RootParams.Boring, "boring", false, "Disable emoji and color output")
	rootCmd.PersistentPreRunE = preRunE
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
