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

var globalOnly bool = true
var configType string
var configFile string

var ErrConfigCreationAborted = fmt.Errorf("configuration file creation aborted")

func promptCreateConfigFileIfNotExist(localPath string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		logger.Error("Failed to determine configuration directory: %v", err)
		return err
	}

	if configFile == "" || (globalOnly && configFile == localPath) {
		if configFile == localPath && !globalOnly {
			logger.Info("Local configuration file already exists and will take precedence: %s", configFile)
		}

		logger.Info("It looks like you haven't run `wow-build-tools config` to set up %s config yet.", configType)
		logger.Prompt("Would you like to create a new %s configuration file? (y/N): ", configType)

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.ToUpper(strings.TrimSpace(response))

		if response == "Y" {
			return createConfigFile(configDir, localPath)
		} else {
			fmt.Println()
			logger.Info("Configuration file creation aborted.")
			return ErrConfigCreationAborted
		}
	}
	return nil
}

func createConfigFile(configDir, localPath string) error {
	if globalOnly {
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
	return nil
}

func setFlavorPath(reader *bufio.Reader, f Flavor, value ...string) error {
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

		if viper.Get("wowPath."+f.Id) != nil {
			defaultPath = viper.GetString("wowPath." + f.Id)
		} else {
			defaultPath += f.SubDir
		}

		logger.Prompt("Enter the path to your %s WoW installation [%s]: ", f.Name, defaultPath)
		flavorPath, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		flavorPath = strings.TrimSpace(flavorPath)
		if len(flavorPath) == 0 {
			flavorPath = defaultPath
		}
	}

	viper.Set("wowPath."+f.Id, flavorPath)
	logger.Success("%s World of Warcraft installation path set to: %s", f.Name, flavorPath)

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
		if !entry.IsDir() {
			continue
		}

		f := flavor.FromDir(entry.Name())

		if f == flavor.UnknownFlavor {
			logger.Warn("Found unknown directory: %s", entry.Name())
			continue
		}

		logger.Success("Found %s World of Warcraft installation at: %s", f.Name, filepath.Join(wowPath, entry.Name()))
		configKey := fmt.Sprintf("wowPath.%s", f.Id)
		viper.Set(configKey, filepath.Join(wowPath, entry.Name()))
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
			for _, f := range flavor.KnownFlavors {
				logger.Info("%d. Update %s World of Warcraft installation path", nextNum, f.Name)
				numberFlavorMap[nextNum] = f
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
	if globalOnly {
		configType = "global"
	}
	configFile = viper.ConfigFileUsed()

	if err = promptCreateConfigFileIfNotExist(localPath); err != nil {
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
			for _, f := range flavor.KnownFlavors {
				validSecondaryArgs = append(validSecondaryArgs, f.Id)
			}
			if slices.Contains(validSecondaryArgs, args[1]) {
				if args[1] == "base" {
					return setWoWPath(bufio.NewReader(os.Stdin), args[2:]...)
				} else {
					f := flavor.FromId(args[1])
					return setFlavorPath(bufio.NewReader(os.Stdin), f, args[2:]...)
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
