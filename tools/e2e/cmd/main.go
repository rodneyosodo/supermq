// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	"github.com/mainflux/mainflux/tools/e2e"
	"github.com/spf13/cobra"
)

func main() {
	econf := e2e.Config{}

	var rootCmd = &cobra.Command{
		Use:   "e2e",
		Short: "e2e is end to end testing tool for Mainflux",
		Long: `Tool for testing end to end flow ow mainflux by creating groups and users and assigning the together andcreating channels and things and connecting them together.
Complete documentation is available at https://docs.mainflux.io`,
		Run: func(_ *cobra.Command, _ []string) {
			e2e.Test(econf)
		},
	}

	// Root Flags
	rootCmd.PersistentFlags().StringVarP(&econf.Host, "host", "", "https://localhost", "address for mainflux instance")
	rootCmd.PersistentFlags().StringVarP(&econf.Prefix, "prefix", "", "", "name prefix for users, groups, things and channels")
	rootCmd.PersistentFlags().StringVarP(&econf.Username, "username", "u", "", "mainflux user")
	rootCmd.PersistentFlags().StringVarP(&econf.Password, "password", "p", "", "mainflux users password")
	rootCmd.PersistentFlags().IntVarP(&econf.Num, "num", "", 10, "number of users, groups, channels and things to create and connect")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
