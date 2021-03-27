/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"bufio"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

var cfgFile string

var ProgramName = "fsnotify-exec"

var rootCmd = &cobra.Command{
	Use: ProgramName,
	RunE: func(cmd *cobra.Command, args []string) error {

		cmd.SilenceUsage = true
		cmd.SilenceErrors = false

		if len(args) > 0 {

			// read commands
			tmpCmd := exec.Command(args[0], args[1:]...)
			tmpCmd.Env = os.Environ()
			tmpCmd.Env = append(tmpCmd.Env, "MY_VAR=some_value")
			out, err := tmpCmd.CombinedOutput()
			if err != nil {
				zap.S().Fatalf("cmd.Run() failed with %s\n", err)
			}
			zap.S().Infof("combined out:\n%s\n", string(out))

			return nil
		}

		// read from pipe
		var sb strings.Builder
		reader := bufio.NewReader(cmd.InOrStdin())

		// try read from pipe
		fileInfo, _ := os.Stdin.Stat()
		if fileInfo.Mode()&os.ModeCharDevice != 0 {

			// no pipe input, no file input, error tips and usage tips
			cmd.SilenceUsage = false
			return errors.New("input command is needed. ")
		}

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

			_, _ = sb.WriteRune(r)
		}

		command := sb.String()
		logrus.Debugf("input command: %s", command)

		tmpCmd := exec.Command("sh", "-c", command)
		tmpCmd.Env = os.Environ()
		tmpCmd.Env = append(tmpCmd.Env, "MY_VAR=some_value")
		out, err := tmpCmd.CombinedOutput()
		if err != nil {
			zap.S().Fatalf("cmd.Run() failed with %s\n", err)
		}
		zap.S().Infof("combined out:\n%s\n", string(out))

		return nil
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		fmt.Sprintf("config file (default is $HOME/.%s/config.toml)", ProgramName),
	)

	rootCmd.Version = version

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(path.Join(home, fmt.Sprintf(".%s", ProgramName)))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err == nil {
		zap.S().Debugf("Using config file:", viper.ConfigFileUsed())
	}
}
