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
	"time"

	"github.com/organicveggie/zfs-snapshot/zfs"
	"github.com/spf13/cobra"
)

var (
	// Recursive enables taking recursive snapshots
	recursive bool

	snapshotCmd = &cobra.Command{
		Use:   "snapshot",
		Short: "Take a ZFS snapshot",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			id := zfs.NewDatasetID(Pool, Dataset)
			snapKey := zfs.MakeSnapshotKey(id, Prefix)

			if err := zfs.TakeSnapshot(snapKey, time.Now(), recursive, DryRun); err != nil {
				fmt.Println(err)
				return
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(snapshotCmd)

	snapshotCmd.Flags().StringVarP(&Pool, "pool", "p", "", "name of the zfs pool")
	snapshotCmd.Flags().StringVarP(&Dataset, "dataset", "d", "", "name of the zfs dataset")
	snapshotCmd.Flags().StringVar(&Prefix, "prefix", "", "snapshot prefix")
	snapshotCmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "enable recursive snapshots")

	snapshotCmd.MarkFlagRequired("pool")
	snapshotCmd.MarkFlagRequired("dataset")
}
