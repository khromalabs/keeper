package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	. "khromalabs/keeper/internal/log"
)

type Config struct {
	Storage string
	Path    map[string]string
	Editor  string
	Ui		string
	Pager	string
	Miniread	bool
}

//go:embed schema.json
var configJsonSchema string

var initerr error
var config Config

func init() {
	initerr = load(&config)
}

func Check() (error) {
	return initerr
}

func Get() (*Config) {
	return &config
}

func assertDataPath(dataPath *string) error {
	var err error
	*dataPath, err = getDataDir()
	if err != nil {
		return fmt.Errorf("Error getting data directory: %v", err)
	}
	return nil
}

func load(config *Config) (error) {
	var err error
	var dataPath string
	envEditor, _ := os.LookupEnv("EDITOR")
	envPager, exists := os.LookupEnv("PAGER")
	if !exists {
		envPager = "less"
	}
	configFileName, exists := os.LookupEnv("KEEPER_CONFIG_FILE")
	if !exists {
		configDir, err := getConfigDir()
		if err != nil {
			fmt.Printf("Error getting config directory: %v", err)
		}
		configFileName = filepath.Join(configDir, "keeper.yaml")
	}
	templatesPath, exists := os.LookupEnv("KEEPER_TEMPLATES_PATH")
	if !exists {
		if err = assertDataPath(&dataPath); err != nil {
			return err
		}
		templatesPath = filepath.Join(dataPath, "templates")
	}
	dbPath, exists := os.LookupEnv("KEEPER_DB_PATH")
	if !exists {
		if err = assertDataPath(&dataPath); err != nil {
			return err
		}
		dbPath = filepath.Join(dataPath, "keeper.db")
	}
	*config = Config{
		Path: map[string]string{
			"templates": templatesPath + string(os.PathSeparator),
			"db": dbPath,
		},
		Storage: "sqlite",
		Ui: "cli",
		Editor: envEditor,
		Pager: envPager,
		Miniread: false,
	}
	configFile, _ := os.ReadFile(configFileName)
	if len(configFile) > 0 {
		var configYaml map[string]interface{}
		if err = yaml.Unmarshal(configFile, &configYaml); err != nil {
			return fmt.Errorf("Error parsing config file: %v", err)
		}
		configJson, err := json.Marshal(configYaml)
		result, err := gojsonschema.Validate(
			gojsonschema.NewStringLoader(configJsonSchema),
			gojsonschema.NewStringLoader(string(configJson)),
		)
		if err != nil {
			return fmt.Errorf("Error converting config YAML to JSON: %v", err)
		}
		if !result.Valid() {
			errmsg := "Config file is not valid. See errors:\n"
			for _, desc := range result.Errors() {
				errmsg += fmt.Sprintf("- %s\n", desc)
			}
			return fmt.Errorf(errmsg)
		}
		// @TODO File configuration not being applied!
		// config.Editor = configYaml["editor"].(string)
	}
	if storageEnv, exists := os.LookupEnv("KEEPER_STORAGE"); exists {
		config.Storage = storageEnv
	}
	if config.Editor == "" {
		return fmt.Errorf("Editor configuration missing (conf editor value or $EDITOR environment variable required)", err)
	}
	enableDebug, exists := os.LookupEnv("KEEPER_ENABLE_DEBUG")
	if enableDebug == "true" || enableDebug == "1" {
		Debug(true)
	}
	return nil
}

func getDataDir() (string, error) {
	var dataHomePath string
	switch runtime.GOOS {
	case "linux", "darwin":
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("Error getting user home directory: %v", err)
			}
			dataHomePath = filepath.Join(userHome, ".local", "share")
		} else {
			dataHomePath = xdgDataHome
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("Error getting APPDATA environment variable")
		}
		dataHomePath = appData
	default:
		return "", fmt.Errorf("Unsupported operating system: %s", runtime.GOOS)
	}
	return filepath.Join(dataHomePath, "keeper"), nil
}


func getConfigDir() (string, error) {
	var configHomePath string
	switch runtime.GOOS {
	case "linux", "darwin":
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("Error getting user config directory: %v", err)
			}
			configHomePath = filepath.Join(userHome, ".config")
		} else {
			configHomePath = xdgConfigHome
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("Error getting APPDATA environment variable")
		}
		configHomePath = appData
	default:
		return "", fmt.Errorf("Unsupported operating system: %s", runtime.GOOS)
	}
	return configHomePath, nil
}

func getTemplatesDir() (string, error) {
	var dataHomePath string
	switch runtime.GOOS {
	case "linux", "darwin":
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("Error getting user home directory: %v", err)
			}
			dataHomePath = filepath.Join(userHome, ".local", "share")
		} else {
			dataHomePath = xdgDataHome
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("Error getting APPDATA environment variable")
		}
		dataHomePath = appData
	default:
		return "", fmt.Errorf("Unsupported operating system: %s", runtime.GOOS)
	}
	return filepath.Join(dataHomePath, "keeper", "templates"), nil
}
