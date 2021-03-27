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
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string

var ProgramName = "fsnotify-exec"

var rootCmd = &cobra.Command{
	Use: ProgramName,
	RunE: func(cmd *cobra.Command, args []string) error {

		cmd.SilenceUsage = true
		cmd.SilenceErrors = false

		var commandEntrypoint string = "sh"
		var commandArgs []string

		if len(args) > 0 {

			// read commands
			commandEntrypoint = args[0]
			commandArgs = args[1:]

		} else {

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

			commandArgs = append(commandArgs, "-c")
			commandArgs = append(commandArgs, command)

		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			zap.S().Fatal(err)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					
					zap.S().Info("event:", event)

					tmpCmd := exec.Command(commandEntrypoint, commandArgs...)

					// 复用系统输入输出
					tmpCmd.Stdout = os.Stdout
					tmpCmd.Stderr = os.Stderr
					tmpCmd.Env = os.Environ()

					// 自定义变量
					tmpCmd.Env = append(tmpCmd.Env, fmt.Sprintf("EVENT=%s", event))

					// 命令处理
					err := tmpCmd.Run()
					if err != nil {
						zap.S().Fatalf("cmd.Run() failed with %s\n", err)
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					zap.S().Fatal("error:", err)
				}
			}
		}()

		err = watcher.Add("./")
		if err != nil {
			zap.S().Info(err)
		}
		<-done

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
