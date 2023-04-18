// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	cc "github.com/ivanpirog/coloredcobra"
	"github.com/mainflux/mainflux/tools/policies-test"
	"github.com/spf13/cobra"
)

func main() {
	pconf := policies.Config{}

	var rootCmd = &cobra.Command{
		Use:   "policies",
		Short: "policies is end-to-end testing tool for Mainflux",
		Long: "Tool for testing end-to-end flow of mainflux by doing a couple of operations namely:\n" +
			"Complete documentation is available at https://docs.mainflux.io",
		Example: "Here is a simple example of using policies tool.\n" +
			"Use the following commands from the root mainflux directory:\n\n" +
			"go run tools/policies/cmd/main.go\n" +
			"go run tools/policies/cmd/main.go --host 142.93.118.47\n",
		Run: func(_ *cobra.Command, _ []string) {
			policies.Test(pconf)
		},
	}

	cc.Init(&cc.Config{
		RootCmd:       rootCmd,
		Headings:      cc.HiCyan + cc.Bold + cc.Underline,
		CmdShortDescr: cc.Magenta,
		Example:       cc.Italic + cc.Magenta,
		ExecName:      cc.Bold,
		Flags:         cc.HiGreen + cc.Bold,
		FlagsDescr:    cc.Green,
		FlagsDataType: cc.White + cc.Italic,
	})

	// Root Flags
	rootCmd.PersistentFlags().StringVarP(&pconf.Host, "host", "H", "localhost", "address for a running mainflux instance")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
