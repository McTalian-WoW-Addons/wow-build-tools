package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/McTalian/wow-build-tools/internal/flavor"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/spf13/viper"
)

type Flavor = flavor.Flavor

var globalConfig bool = true
var configType string
var configFile string

var ErrConfigCreationAborted = fmt.Errorf("configuration file creation aborted")

func createConfigFileIfNotExist(localPath string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		logger.Error("Failed to determine configuration directory: %v", err)
		return err
	}

	if configFile == "" || (globalConfig && configFile == localPath) {
		if configFile == localPath {
			logger.Info("Local configuration file already exists and will take precedence: %s", configFile)
		}

		logger.Info("It looks like you haven't run `wow-build-tools config` to set up %s config yet.", configType)
		logger.Prompt("Would you like to create a new %s configuration file? (y/N): ", configType)

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(response)

		if response == "y" || response == "Y" {
			if globalConfig {
				logger.Info("Creating global configuration file...")
				err := os.MkdirAll(configDir, 0755)
				if err != nil && !os.IsExist(err) {
					return err
				}
				err = viper.SafeWriteConfig()
				if err != nil {
					return err
				}
				err = viper.ReadInConfig()
				if err != nil {
					return err
				}
				logger.Success("Configuration file created: %s", viper.ConfigFileUsed())
			} else {
				logger.Info("--global flag not set, creating local configuration file...")
				logger.Info("If you want to create a global configuration file instead, run `wow-build-tools config --global`.")
				err := viper.SafeWriteConfigAs(".wbt.yaml")
				if err != nil {
					return err
				}
				logger.Success("Configuration file created: %s", localPath)
			}
		} else {
			fmt.Println()
			logger.Info("Configuration file creation aborted.")
			return ErrConfigCreationAborted
		}
	}
	return nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func setFlavorPath(reader *bufio.Reader, flavor Flavor, value ...string) error {
	var flavorPath string
	var err error
	if len(value) == 1 {
		flavorPath = value[0]
	} else {
		basePath := viper.Get("wowPath.base")
		var defaultPath string
		if basePath != nil {
			defaultPath = viper.GetString("wowPath.base") + string(os.PathSeparator)
		} else if os.PathSeparator == '\\' {
			defaultPath = "C:\\Program Files (x86)\\World of Warcraft\\"
		} else {
			defaultPath = "/mnt/c/Program Files (x86)/World of Warcraft/"
		}

		if viper.Get("wowPath."+string(flavor)) != nil {
			defaultPath = viper.GetString("wowPath." + string(flavor))
		} else {
			defaultPath += flavor.ToDir()
		}

		logger.Prompt("Enter the path to your %s WoW installation [%s]: ", capitalize(string(flavor)), defaultPath)
		flavorPath, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		flavorPath = strings.TrimSpace(flavorPath)
		if len(flavorPath) == 0 {
			flavorPath = defaultPath
		}
	}

	viper.Set("wowPath."+string(flavor), flavorPath)
	logger.Success("%s World of Warcraft installation path set to: %s", capitalize(string(flavor)), flavorPath)

	return nil
}

func setWoWPath(reader *bufio.Reader, value ...string) error {
	var wowPath string
	var err error
	if len(value) == 1 {
		wowPath = value[0]
	} else {
		var defaultPath string
		if viper.Get("wowPath.base") != nil {
			defaultPath = viper.GetString("wowPath.base")
			// Check if Windows or Unix
		} else if os.PathSeparator == '\\' {
			defaultPath = "C:\\Program Files (x86)\\World of Warcraft"
		} else {
			defaultPath = "/mnt/c/Program Files (x86)/World of Warcraft"
		}

		logger.Prompt("Enter the path to your WoW installation [%s]: ", defaultPath)
		wowPath, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		wowPath = strings.TrimSpace(wowPath)
		if wowPath == "" {
			wowPath = defaultPath
		}
	}
	viper.Set("wowPath.base", wowPath)
	logger.Success("World of Warcraft installation path set to: %s", wowPath)

	fmt.Println()

	contents, err := os.ReadDir(wowPath)
	if err != nil {
		logger.Warn("Could not read directory: %v", err)
		logger.Warn("Please make sure the path is correct and try again.")
		return err
	}

	for _, entry := range contents {
		if entry.IsDir() && entry.Name() == flavor.Retail.ToDir() {
			logger.Success("Found Retail World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.retail", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.Classic.ToDir() {
			logger.Success("Found Classic World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.classic", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.ClassicEra.ToDir() {
			logger.Success("Found Classic Era World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.classicEra", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.ClassicEraPtr.ToDir() {
			logger.Success("Found Classic Era PTR World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.classicEraPtr", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.Ptr.ToDir() {
			logger.Success("Found PTR World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.ptr", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.Xptr.ToDir() {
			logger.Success("Found XPTR World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.xptr", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.ClassicPtr.ToDir() {
			logger.Success("Found Classic PTR World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.classicPtr", filepath.Join(wowPath, entry.Name()))
		} else if entry.IsDir() && entry.Name() == flavor.ClassicBeta.ToDir() {
			logger.Success("Found Classic Beta World of Warcraft installation at: %s", filepath.Join(wowPath, entry.Name()))
			viper.Set("wowPath.classicBeta", filepath.Join(wowPath, entry.Name()))
		} else {
			logger.Warn("Found unknown directory: %s", entry.Name())
		}
	}

	wowPaths := viper.GetStringMapString("wowPath")
	if len(wowPaths) == 1 {
		logger.Error("No valid World of Warcraft installations found in the specified path.")
		logger.Error("Please make sure the path is correct and try again.")
		return fmt.Errorf("no valid World of Warcraft installations found in the specified path")
	} else {
		logger.Success("World of Warcraft installation paths set successfully!")
	}
	return nil
}

func runConfigWizard() error {
	reader := bufio.NewReader(os.Stdin)

	logger.Info("Welcome to the wow-build-tools configuration wizard!")
	logger.Info("Please follow the prompts to set up your configuration.")

	var numberFlavorMap = map[int]Flavor{}

	for {
		wowPaths := viper.GetStringMapString("wowPath")
		logger.Info("\nConfiguration Menu:")
		logger.Info("1. Set or update Base World of Warcraft installation path")
		nextNum := 2
		if len(wowPaths) >= 1 {
			for _, flavor := range flavor.KnownFlavors {
				logger.Info("%d. Update %s World of Warcraft installation path", nextNum, capitalize(string(flavor)))
				numberFlavorMap[nextNum] = flavor
				nextNum++
			}
		}
		logger.Info("%d. Save and exit", nextNum)
		nextNum++
		logger.Info("%d. Exit without saving", nextNum)
		logger.Prompt("Enter your choice: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err = setWoWPath(reader)
			if err != nil {
				return err
			}
		case strconv.Itoa(nextNum - 1):
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
			logger.Success("Configuration saved successfully!")
			return nil
		case strconv.Itoa(nextNum):
			logger.Info("Exiting without saving.")
			return nil
		default:
			if len(wowPaths) >= 1 {
				choiceNum, err := strconv.Atoi(choice)
				if err != nil {
					logger.Warn("Invalid choice, please try again.")
				} else if flavor, ok := numberFlavorMap[choiceNum]; ok {
					err = setFlavorPath(reader, flavor)
					if err != nil {
						return err
					}
				} else {
					logger.Warn("Invalid choice, please try again.")
				}
			} else {
				logger.Warn("Invalid choice, please try again.")
			}
		}
	}
}

func RunConfig(args []string) error {
	localPath, err := filepath.Abs(filepath.Join(".", ".wbt.yaml"))
	if err != nil {
		return err
	}

	configType = "local"
	if globalConfig {
		configType = "global"
	}
	configFile = viper.ConfigFileUsed()

	if err = createConfigFileIfNotExist(localPath); err != nil {
		if err == ErrConfigCreationAborted {
			return nil
		}
		return err
	}

	validPrimaryArgs := []string{"wowPath"}
	if len(args) > 0 {
		if slices.Contains(validPrimaryArgs, args[0]) {
			if len(args) == 1 {
				logger.Error("No subcommand provided for %s configuration", args[0])
				return fmt.Errorf("no subcommand provided for %s configuration", args[0])
			}
			validSecondaryArgs := []string{"base"}
			for _, flavor := range flavor.KnownFlavors {
				validSecondaryArgs = append(validSecondaryArgs, string(flavor))
			}
			if slices.Contains(validSecondaryArgs, args[1]) {
				if args[1] == "base" {
					return setWoWPath(bufio.NewReader(os.Stdin), args[2:]...)
				} else {
					flavor := Flavor(args[1])
					return setFlavorPath(bufio.NewReader(os.Stdin), flavor, args[2:]...)
				}
			} else {
				logger.Error("Invalid subcommand provided for %s configuration, %s. Must be one of %v", args[0], args[1], validSecondaryArgs)
				return fmt.Errorf("invalid subcommand provided for %s configuration", args[0])
			}
		} else {
			logger.Error("Invalid primary argument provided for configuration, %s. Must be one of %v", args[0], validPrimaryArgs)
			return fmt.Errorf("invalid primary argument provided for configuration")
		}
	}

	return runConfigWizard()
}
