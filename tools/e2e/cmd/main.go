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
		Long: `Tool for testing end to end flow of mainflux by creating groups and users and assigning the together andcreating channels and things and connecting them together.
Complete documentation is available at https://docs.mainflux.io`,
		Run: func(_ *cobra.Command, _ []string) {
			e2e.Test(econf)
		},
	}

	// Root Flags
	rootCmd.PersistentFlags().StringVarP(&econf.Host, "host", "H", "localhost", "address for a running mainflux instance")
	rootCmd.PersistentFlags().StringVarP(&econf.Prefix, "prefix", "p", "", "name prefix for users, groups, things and channels")
	rootCmd.PersistentFlags().Uint64VarP(&econf.Num, "num", "n", uint64(10), "number of users, groups, channels and things to create and connect")
	rootCmd.PersistentFlags().Uint64VarP(&econf.NumOfMsg, "num_of_messages", "N", uint64(10), "number of messages to send")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
