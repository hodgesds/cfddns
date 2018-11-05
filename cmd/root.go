// Copyright © 2018 Daniel Hodges
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hodgesds/cfddns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cfddns",
	Short: "Cloudflare Dynamic DNS daemon",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := viper.GetViper()

		api, err := cloudflare.New(cfg.GetString("key"), cfg.GetString("email"))
		if err != nil {
			log.Fatal(err)
		}

		zoneID, err := api.ZoneIDByName(cfg.GetString("domain"))
		if err != nil {
			log.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go cfddns.Daemon(ctx, zoneID, api, cfg.GetDuration("interval"))
		<-sigChan
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config", "",
		"config file (default is $HOME/.cfddns.yaml)",
	)
	RootCmd.PersistentFlags().DurationP(
		"interval", "i",
		1*time.Hour,
		"DDNS refresh interval",
	)
	RootCmd.PersistentFlags().StringP(
		"domain", "d",
		"",
		"DDNS domain",
	)
	RootCmd.PersistentFlags().StringP(
		"key", "k",
		"",
		"Cloudflare API key",
	)
	RootCmd.PersistentFlags().StringP(
		"email", "e",
		"",
		"Cloudflare API email",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.BindPFlags(RootCmd.PersistentFlags())
	viper.SetConfigName(".cfddns")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
