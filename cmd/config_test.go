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

	// Test case 1: Repository config should be found first
	t.Run("repository config found first", func(t *testing.T) {
		// Create repository config
		repoConfigPath := filepath.Join(repoDir, ".gitallica.yml")
		repoConfig := `test_setting: "repo_value"
author_mappings:
  - patterns: ["repo"]
    canonical: "repo@example.com"`
		err = os.WriteFile(repoConfigPath, []byte(repoConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write repo config: %v", err)
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

		// Reset viper and test config loading
		viper.Reset()

		// Test the config loading logic
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gitallica")

		// Should find and load repository config
		err = viper.ReadInConfig()
		if err != nil {
			t.Fatalf("Failed to read repository config: %v", err)
		}

		// Verify repository config was loaded
		if viper.GetString("test_setting") != "repo_value" {
			t.Errorf("Expected test_setting to be 'repo_value', got '%s'", viper.GetString("test_setting"))
		}
	})

	// Test case 2: Global config should be used when no repository config exists
	t.Run("global config fallback", func(t *testing.T) {
		// Create global config
		globalConfigPath := filepath.Join(homeDir, ".gitallica.yaml")
		globalConfig := `test_setting: "global_value"
author_mappings:
  - patterns: ["global"]
    canonical: "global@example.com"`
		err = os.WriteFile(globalConfigPath, []byte(globalConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write global config: %v", err)
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

		// Test the config loading logic - first try repository config
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gitallica")

		// Should not find repository config
		err = viper.ReadInConfig()
		if err == nil {
			t.Error("Expected error when reading non-existent repository config")
		}

		// Now try global config
		viper.Reset()
		viper.AddConfigPath(homeDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gitallica")

		// Should find and load global config
		err = viper.ReadInConfig()
		if err != nil {
			t.Fatalf("Failed to read global config: %v", err)
		}

		// Verify global config was loaded
		if viper.GetString("test_setting") != "global_value" {
			t.Errorf("Expected test_setting to be 'global_value', got '%s'", viper.GetString("test_setting"))
		}
	})

	// Test case 3: Explicit config file should override both
	t.Run("explicit config file override", func(t *testing.T) {
		// Create explicit config file
		explicitConfigPath := filepath.Join(tempDir, "explicit.yaml")
		explicitConfig := `test_setting: "explicit_value"
author_mappings:
  - patterns: ["explicit"]
    canonical: "explicit@example.com"`
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