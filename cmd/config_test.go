package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRepositoryConfigHierarchy(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "gitallica-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectory to simulate repository
	repoDir := filepath.Join(tempDir, "repo")
	err = os.Mkdir(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Create home directory
	homeDir := filepath.Join(tempDir, "home")
	err = os.Mkdir(homeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create home dir: %v", err)
	}

	// Test case 1: Project config overrides home config
	t.Run("project config overrides home config", func(t *testing.T) {
		// Create home config
		homeConfigPath := filepath.Join(homeDir, ".gitallica.yaml")
		homeConfig := `test_setting: "home_value"
global_setting: "home_global"
churn:
  paths:
    - "home/path1"
    - "home/path2"`
		err = os.WriteFile(homeConfigPath, []byte(homeConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write home config: %v", err)
		}

		// Create project config
		projectConfigPath := filepath.Join(repoDir, ".gitallica.yml")
		projectConfig := `test_setting: "project_value"
project_setting: "project_specific"
churn:
  paths:
    - "project/path1"
    - "project/path2"`
		err = os.WriteFile(projectConfigPath, []byte(projectConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write project config: %v", err)
		}

		// Change to repo directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current dir: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(repoDir)
		if err != nil {
			t.Fatalf("Failed to change to repo dir: %v", err)
		}

		// Reset viper and test config loading with hierarchy
		viper.Reset()

		// Mock the home directory for testing
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		defer os.Setenv("HOME", originalHome)

		// Call the actual initConfig function
		initConfig()

		// Verify project config overrides home config
		if viper.GetString("test_setting") != "project_value" {
			t.Errorf("Expected test_setting to be 'project_value', got '%s'", viper.GetString("test_setting"))
		}

		// Verify project-specific setting is available
		if viper.GetString("project_setting") != "project_specific" {
			t.Errorf("Expected project_setting to be 'project_specific', got '%s'", viper.GetString("project_setting"))
		}

		// Verify home-only setting is still available
		if viper.GetString("global_setting") != "home_global" {
			t.Errorf("Expected global_setting to be 'home_global', got '%s'", viper.GetString("global_setting"))
		}

		// Verify project config overrides home config for nested settings
		projectPaths := viper.GetStringSlice("churn.paths")
		expectedPaths := []string{"project/path1", "project/path2"}
		if len(projectPaths) != len(expectedPaths) {
			t.Errorf("Expected %d project paths, got %d", len(expectedPaths), len(projectPaths))
		}
		for i, path := range expectedPaths {
			if i < len(projectPaths) && projectPaths[i] != path {
				t.Errorf("Expected project path %d to be '%s', got '%s'", i, path, projectPaths[i])
			}
		}
	})

	// Test case 2: Home config fallback when no project config exists
	t.Run("home config fallback", func(t *testing.T) {
		// Create home config
		homeConfigPath := filepath.Join(homeDir, ".gitallica.yaml")
		homeConfig := `test_setting: "home_value"
global_setting: "home_global"`
		err = os.WriteFile(homeConfigPath, []byte(homeConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write home config: %v", err)
		}

		// Create empty repo directory (no config file)
		emptyRepoDir := filepath.Join(tempDir, "empty-repo")
		err = os.Mkdir(emptyRepoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty repo dir: %v", err)
		}

		// Change to empty repo directory
		err = os.Chdir(emptyRepoDir)
		if err != nil {
			t.Fatalf("Failed to change to empty repo dir: %v", err)
		}

		// Reset viper and test config loading
		viper.Reset()

		// Mock the home directory for testing
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		defer os.Setenv("HOME", originalHome)

		// Call the actual initConfig function
		initConfig()

		// Verify home config was loaded
		if viper.GetString("test_setting") != "home_value" {
			t.Errorf("Expected test_setting to be 'home_value', got '%s'", viper.GetString("test_setting"))
		}
		if viper.GetString("global_setting") != "home_global" {
			t.Errorf("Expected global_setting to be 'home_global', got '%s'", viper.GetString("global_setting"))
		}
	})

	// Test case 3: Explicit config file should override both
	t.Run("explicit config file override", func(t *testing.T) {
		// Create explicit config file
		explicitConfigPath := filepath.Join(tempDir, "explicit.yaml")
		explicitConfig := `test_setting: "explicit_value"
explicit_setting: "explicit_only"`
		err = os.WriteFile(explicitConfigPath, []byte(explicitConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write explicit config: %v", err)
		}

		// Reset viper and test explicit config loading
		viper.Reset()

		// Set the config file flag
		cfgFile = explicitConfigPath
		defer func() { cfgFile = "" }()

		// Call the actual initConfig function
		initConfig()

		// Verify explicit config was loaded
		if viper.GetString("test_setting") != "explicit_value" {
			t.Errorf("Expected test_setting to be 'explicit_value', got '%s'", viper.GetString("test_setting"))
		}
		if viper.GetString("explicit_setting") != "explicit_only" {
			t.Errorf("Expected explicit_setting to be 'explicit_only', got '%s'", viper.GetString("explicit_setting"))
		}
	})
}

func TestUpwardConfigSearch(t *testing.T) {
	// Test that config files are found in parent directories
	tempDir, err := os.MkdirTemp("", "gitallica-upward-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	parentDir := filepath.Join(tempDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	err = os.MkdirAll(childDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested dirs: %v", err)
	}

	// Create config in parent directory
	parentConfigPath := filepath.Join(parentDir, ".gitallica.yaml")
	parentConfig := `test_setting: "parent_value"
parent_only: "parent_specific"`
	err = os.WriteFile(parentConfigPath, []byte(parentConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Change to child directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(childDir)
	if err != nil {
		t.Fatalf("Failed to change to child dir: %v", err)
	}

	// Reset viper and test config loading
	viper.Reset()

	// Call the actual initConfig function
	initConfig()

	// Verify parent config was found and loaded
	if viper.GetString("test_setting") != "parent_value" {
		t.Errorf("Expected test_setting to be 'parent_value', got '%s'", viper.GetString("test_setting"))
	}
	if viper.GetString("parent_only") != "parent_specific" {
		t.Errorf("Expected parent_only to be 'parent_specific', got '%s'", viper.GetString("parent_only"))
	}
}

func TestConfigFileExtensions(t *testing.T) {
	// Test that both .yml and .yaml extensions work
	tempDir, err := os.MkdirTemp("", "gitallica-config-ext-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name     string
		filename string
	}{
		{"yaml extension", ".gitallica.yaml"},
		{"yml extension", ".gitallica.yml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, tc.filename)
			config := `test_setting: "test_value"`
			err = os.WriteFile(configPath, []byte(config), 0644)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Change to temp directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current dir: %v", err)
			}
			defer os.Chdir(originalDir)

			err = os.Chdir(tempDir)
			if err != nil {
				t.Fatalf("Failed to change to temp dir: %v", err)
			}

			// Reset viper and test config loading
			viper.Reset()

			viper.AddConfigPath(".")
			viper.SetConfigType("yaml")
			viper.SetConfigName(".gitallica")

			err = viper.ReadInConfig()
			if err != nil {
				t.Fatalf("Failed to read config file %s: %v", tc.filename, err)
			}

			if viper.GetString("test_setting") != "test_value" {
				t.Errorf("Expected test_setting to be 'test_value', got '%s'", viper.GetString("test_setting"))
			}
		})
	}
}

func TestGetConfigPathsHierarchy(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "gitallica-paths-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectory to simulate repository
	repoDir := filepath.Join(tempDir, "repo")
	err = os.Mkdir(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Test case 1: Command-specific config takes precedence over defaults
	t.Run("command-specific config overrides defaults", func(t *testing.T) {
		// Create config with both command-specific and default paths
		configPath := filepath.Join(repoDir, ".gitallica.yaml")
		config := `churn:
  paths:
    - "cmd/"
    - "main.go"
    
defaults:
  paths:
    - "app/Module1"
    - "app/Module2"`
		err = os.WriteFile(configPath, []byte(config), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Change to repo directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current dir: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(repoDir)
		if err != nil {
			t.Fatalf("Failed to change to repo dir: %v", err)
		}

		// Reset viper and load config with clean environment
		viper.Reset()

		// Mock empty home directory to avoid loading home config
		originalHome := os.Getenv("HOME")
		emptyHome := filepath.Join(tempDir, "empty-home")
		os.Mkdir(emptyHome, 0755)
		os.Setenv("HOME", emptyHome)
		defer os.Setenv("HOME", originalHome)

		initConfig()

		// Create mock command
		mockCmd := &cobra.Command{}
		mockCmd.Flags().StringSlice("path", []string{}, "test path flag")

		// Test getConfigPaths
		paths, source := getConfigPaths(mockCmd, "churn.paths")
		expectedPaths := []string{"cmd/", "main.go"}

		if len(paths) != len(expectedPaths) {
			t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
		}

		for i, expected := range expectedPaths {
			if i < len(paths) && paths[i] != expected {
				t.Errorf("Expected path %d to be '%s', got '%s'", i, expected, paths[i])
			}
		}

		if source != "(from config)" {
			t.Errorf("Expected source to be '(from config)', got '%s'", source)
		}
	})

	// Test case 2: Defaults used when no command-specific config exists
	t.Run("defaults used when no command-specific config", func(t *testing.T) {
		// Create config with only defaults
		configPath := filepath.Join(repoDir, ".gitallica.yaml")
		config := `defaults:
  paths:
    - "src/modules"
    - "lib/"`
		err = os.WriteFile(configPath, []byte(config), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Change to repo directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current dir: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(repoDir)
		if err != nil {
			t.Fatalf("Failed to change to repo dir: %v", err)
		}

		// Reset viper and load config with clean environment
		viper.Reset()

		// Mock empty home directory to avoid loading home config
		originalHome := os.Getenv("HOME")
		emptyHome := filepath.Join(tempDir, "empty-home2")
		os.Mkdir(emptyHome, 0755)
		os.Setenv("HOME", emptyHome)
		defer os.Setenv("HOME", originalHome)

		initConfig()

		// Create mock command
		mockCmd := &cobra.Command{}
		mockCmd.Flags().StringSlice("path", []string{}, "test path flag")

		// Test getConfigPaths for a command without specific config
		paths, source := getConfigPaths(mockCmd, "bus-factor.paths")
		expectedPaths := []string{"src/modules", "lib/"}

		if len(paths) != len(expectedPaths) {
			t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
		}

		for i, expected := range expectedPaths {
			if i < len(paths) && paths[i] != expected {
				t.Errorf("Expected path %d to be '%s', got '%s'", i, expected, paths[i])
			}
		}

		if source != "(from defaults)" {
			t.Errorf("Expected source to be '(from defaults)', got '%s'", source)
		}
	})

	// Test case 3: CLI flags override everything
	t.Run("CLI flags override config and defaults", func(t *testing.T) {
		// Create config with both command-specific and default paths
		configPath := filepath.Join(repoDir, ".gitallica.yaml")
		config := `churn:
  paths:
    - "cmd/"
    
defaults:
  paths:
    - "app/"`
		err = os.WriteFile(configPath, []byte(config), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Change to repo directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current dir: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(repoDir)
		if err != nil {
			t.Fatalf("Failed to change to repo dir: %v", err)
		}

		// Reset viper and load config with clean environment
		viper.Reset()

		// Mock empty home directory to avoid loading home config
		originalHome := os.Getenv("HOME")
		emptyHome := filepath.Join(tempDir, "empty-home3")
		os.Mkdir(emptyHome, 0755)
		os.Setenv("HOME", emptyHome)
		defer os.Setenv("HOME", originalHome)

		initConfig()

		// Create mock command with CLI flags set
		mockCmd := &cobra.Command{}
		mockCmd.Flags().StringSlice("path", []string{}, "test path flag")

		// Set CLI flag
		err = mockCmd.Flags().Set("path", "test/")
		if err != nil {
			t.Fatalf("Failed to set CLI path flag: %v", err)
		}

		// Test getConfigPaths
		paths, source := getConfigPaths(mockCmd, "churn.paths")
		expectedPaths := []string{"test/"}

		if len(paths) != len(expectedPaths) {
			t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
		}

		if len(paths) > 0 && paths[0] != expectedPaths[0] {
			t.Errorf("Expected path to be '%s', got '%s'", expectedPaths[0], paths[0])
		}

		if source != "(from CLI)" {
			t.Errorf("Expected source to be '(from CLI)', got '%s'", source)
		}
	})

	// Test case 4: Empty result when no paths configured anywhere
	t.Run("empty result when no paths configured", func(t *testing.T) {
		// Create empty config
		configPath := filepath.Join(repoDir, ".gitallica.yaml")
		config := `# Empty config`
		err = os.WriteFile(configPath, []byte(config), 0644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Change to repo directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current dir: %v", err)
		}
		defer os.Chdir(originalDir)

		err = os.Chdir(repoDir)
		if err != nil {
			t.Fatalf("Failed to change to repo dir: %v", err)
		}

		// Reset viper and load config with clean environment
		viper.Reset()

		// Mock empty home directory to avoid loading home config
		originalHome := os.Getenv("HOME")
		emptyHome := filepath.Join(tempDir, "empty-home4")
		os.Mkdir(emptyHome, 0755)
		os.Setenv("HOME", emptyHome)
		defer os.Setenv("HOME", originalHome)

		initConfig()

		// Create mock command
		mockCmd := &cobra.Command{}
		mockCmd.Flags().StringSlice("path", []string{}, "test path flag")

		// Test getConfigPaths
		paths, source := getConfigPaths(mockCmd, "nonexistent.paths")

		if len(paths) != 0 {
			t.Errorf("Expected 0 paths, got %d", len(paths))
		}

		if source != "" {
			t.Errorf("Expected empty source, got '%s'", source)
		}
	})
}
