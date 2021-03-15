/*
Copyright © 2021 Sean Laurent <o r g a n i c v e g g i e @ Google Mail>

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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Verbose indicates whether or not to perform verbose logging.
	Verbose bool

	// DryRun indicates whether or not to execute any commands.
	DryRun bool

	// Pool is the name of the ZFS pool.
	Pool string

	// Dataset is the name fo the ZFS dataset within the pool.
	Dataset string

	// Optional prefix to use when creating / destroying snapshots.
	Prefix string

	rootCmd = &cobra.Command{
		Use:   "zfs-snapshot",
		Short: "Manages ZFS snapshots",
		Long:  ``,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&DryRun, "dry-run", false, "dry-run mode")
}

// Execute executes the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}
