/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitclone <Git Repo>",
	Short: "Clones a git repo and attempt to set it up",
	Long: `Clone a git repo and attempt to set it up.
  
  Go (mod): ` + "`go get`" + `
  NPM:      ` + "`npm i`" + `
  Yarn:     ` + "`yarn`" + `
  
  No matter the project type, the binary presence will be validated first.`,
	Args: cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		execCmd("git", ".", append([]string{"clone"}, args...)...)
		lastArg := args[len(args)-1]
		split := strings.Split(lastArg, "/")

		dir := split[len(split)-1]
		dir = strings.ReplaceAll(dir, ".git", "")

		pType := projectType(dir)
		if pType == "" {
			fmt.Printf("No recognizable project in %s\n", dir)
			os.Exit(0)
		} else {
			fmt.Printf("Found %s project in %s\n", pType, dir)
			if pType == "go" {
				fmt.Printf("Executing `go get` in %s\n", dir)
				execCmd("go", dir, "get")
			} else if pType == "yarn" {
				fmt.Printf("Executing `yarn` in %s\n", dir)
				execCmd("yarn", dir)
			} else if pType == "npm" {
				fmt.Printf("Executing `npm i` in %s\n", dir)
				execCmd("npm", dir, "i")
			}
		}
	},
}

func projectType(dir string) string {
	if fileExists("./" + dir + "/go.mod") {
		return "go"
	}

	if fileExists("./" + dir + "/yarn.lock") {
		return "yarn"
	}

	if fileExists("./" + dir + "/package-lock.json") {
		return "npm"
	}

	return ""
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func execCmd(cmdStr, dir string, args ...string) {
	_, err := exec.LookPath(cmdStr)
	if err != nil {
		log.Fatalf("%s command not found", cmdStr)
	}

	cmd := exec.Command(cmdStr, args...)
	cmd.Dir = dir

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	if errStdout != nil || errStderr != nil {
		log.Fatalln("failed to capture stdout of stderr")
	}

	// outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	// fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gitclone.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gitclone" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gitclone")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
