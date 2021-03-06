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
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var cfgFile string

var ProgramName = "fsnotify-exec"

var watchedObjects, ignoredGlobPatterns, ignoredRegexPatterns []string

var rootCmd = &cobra.Command{
	Use: ProgramName,
	RunE: func(cmd *cobra.Command, args []string) error {

		cmd.SilenceUsage = true
		cmd.SilenceErrors = false

		var commandEntrypoint string = "sh"
		var commandArgs []string

		var sb strings.Builder
		reader := bufio.NewReader(cmd.InOrStdin())

		if len(args) > 0 {

			// read commands

			commandArgs = append(commandArgs, "-c")
			for _, arg := range args {
				sb.WriteString(fmt.Sprintf("%s ", arg))
			}
			commandArgs = append(commandArgs, gstr.Trim(sb.String()))

		} else {

			// read from pipe

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

					// 忽略规则内的文件
					for _, pattern := range ignoredRegexPatterns {
						if gregex.IsMatchString(pattern, event.Name) {
							return
						}
					}

					tmpCmd := exec.Command(commandEntrypoint, commandArgs...)

					// 复用系统输入输出
					tmpCmd.Env = os.Environ()

					// 自定义变量
					tmpCmd.Env = append(tmpCmd.Env, fmt.Sprintf("NOTIFY_EVENT=%s", event.Op))
					tmpCmd.Env = append(tmpCmd.Env, fmt.Sprintf("NOTIFY_FILE=%s", event.Name))

					zap.S().Debug("//////////////////////////////////////")
					zap.S().Debugf("NOTIFY_FILE: %s", event.Name)
					zap.S().Debugf("NOTIFY_EVENT: %s", event.Op)
					zap.S().Debugf("Env: %s", tmpCmd.Env)
					zap.S().Debugf("commandEntrypoint: %s", commandEntrypoint)
					zap.S().Debugf("commandArgs: %s", commandArgs)

					// 命令处理
					output, err := tmpCmd.CombinedOutput()
					if err != nil {
						zap.S().Fatalf("cmd.Run() failed with %s\n", err)
					}
					fmt.Println(gconv.String(output))

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					zap.S().Fatal("error:", err)
				}
			}
		}()

		// 新增匹配
		for _, obj := range watchedObjects {
			if err := watcher.Add(obj); err != nil {
				zap.S().Info(err)
			}
		}

		// 新增忽略(Glob)
		for _, pattern := range ignoredGlobPatterns {

			list, err := filepath.Glob(pattern)
			if err != nil {
				zap.S().Info(err)
				continue
			}
			for _, obj := range list {
				if err := watcher.Remove(obj); err != nil {
					zap.S().Info(err)
				}
			}

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

	rootCmd.Flags().StringSliceVarP(&watchedObjects, "watch", "w", []string{"./"}, "thd object which need be watched, eg: dir/file.")

	rootCmd.Flags().StringSliceVarP(&ignoredGlobPatterns, "ignore-glob", "", []string{}, "thd object which need be ignored(glob/wild).")

	rootCmd.Flags().StringSliceVarP(&ignoredRegexPatterns, "ignore-regex", "", []string{}, "thd object which need be ignored(regex).")

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
