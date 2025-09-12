/*
Copyright © 2025 Ben Ricker

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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitallica",
	Short: "Shred your git history. Rock your repo insights.",
	Long: `gitallica performs temporal diff analysis of distributed version control logs
to help you understand code evolution, identify risks, and optimize team workflows.

Analyze churn patterns, code survival rates, and other engineering metrics
to make data-driven decisions about your codebase health.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gitallica.yaml)")
}

// initConfig reads in config file with proper hierarchy:
// 1. Explicit config file (--config flag) - highest priority
// 2. Project-specific .gitallica.yaml/.gitallica.yml in current directory
// 3. Home directory ~/.gitallica.yaml - lowest priority
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag (highest priority).
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
		return
	}

	// Configuration hierarchy: project-specific overrides home directory
	// First, try to load home directory config as base
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Load home config first (if it exists)
	homeViper := viper.New()
	homeViper.AddConfigPath(home)
	homeViper.SetConfigType("yaml")
	homeViper.SetConfigName(".gitallica")
	
	if err := homeViper.ReadInConfig(); err == nil {
		// Merge home config into main viper
		for _, key := range homeViper.AllKeys() {
			viper.Set(key, homeViper.Get(key))
		}
	}

	// Then try to load project-specific config (overrides home config)
	projectViper := viper.New()
	projectViper.AddConfigPath(".")
	projectViper.SetConfigType("yaml")
	projectViper.SetConfigName(".gitallica")
	
	if err := projectViper.ReadInConfig(); err == nil {
		// Merge project config into main viper (overrides home config)
		for _, key := range projectViper.AllKeys() {
			viper.Set(key, projectViper.Get(key))
		}
		fmt.Fprintln(os.Stderr, "Using project config file:", projectViper.ConfigFileUsed())
	} else if homeViper.ConfigFileUsed() != "" {
		fmt.Fprintln(os.Stderr, "Using home config file:", homeViper.ConfigFileUsed())
	}
}
