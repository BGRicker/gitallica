package cmd

import (
	"os"
	"path/filepath"
	"testing"

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

		// Simulate the new initConfig logic
		// Load home config first
		homeViper := viper.New()
		homeViper.AddConfigPath(homeDir)
		homeViper.SetConfigType("yaml")
		homeViper.SetConfigName(".gitallica")
		
		if err := homeViper.ReadInConfig(); err == nil {
			for _, key := range homeViper.AllKeys() {
				viper.Set(key, homeViper.Get(key))
			}
		}

		// Load project config (overrides home config)
		projectViper := viper.New()
		projectViper.AddConfigPath(".")
		projectViper.SetConfigType("yaml")
		projectViper.SetConfigName(".gitallica")
		
		if err := projectViper.ReadInConfig(); err == nil {
			for _, key := range projectViper.AllKeys() {
				viper.Set(key, projectViper.Get(key))
			}
		}

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

		// Simulate the new initConfig logic
		// Load home config first
		homeViper := viper.New()
		homeViper.AddConfigPath(homeDir)
		homeViper.SetConfigType("yaml")
		homeViper.SetConfigName(".gitallica")
		
		if err := homeViper.ReadInConfig(); err == nil {
			for _, key := range homeViper.AllKeys() {
				viper.Set(key, homeViper.Get(key))
			}
		}

		// Try to load project config (should fail)
		projectViper := viper.New()
		projectViper.AddConfigPath(".")
		projectViper.SetConfigType("yaml")
		projectViper.SetConfigName(".gitallica")
		
		// Should not find project config
		err = projectViper.ReadInConfig()
		if err == nil {
			t.Error("Expected error when reading non-existent project config")
		}

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

		// Test explicit config loading
		viper.SetConfigFile(explicitConfigPath)

		err = viper.ReadInConfig()
		if err != nil {
			t.Fatalf("Failed to read explicit config: %v", err)
		}

		// Verify explicit config was loaded
		if viper.GetString("test_setting") != "explicit_value" {
			t.Errorf("Expected test_setting to be 'explicit_value', got '%s'", viper.GetString("test_setting"))
		}
		if viper.GetString("explicit_setting") != "explicit_only" {
			t.Errorf("Expected explicit_setting to be 'explicit_only', got '%s'", viper.GetString("explicit_setting"))
		}
	})
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