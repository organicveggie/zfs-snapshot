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
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/organicveggie/zfs-snapshot/zfs"
	"github.com/spf13/cobra"
)

const (
	oneDay  = 24 * time.Hour
	oneWeek = 24 * 7 * time.Hour

	snapTimeFmt = "20060102150405"
)

var (
	saveCount    int
	saveDuration time.Duration

	// cleanupCmd represents the cleanup command
	cleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Cleans up existing snapshots",
		Long:  `Removes existing snapshots meeting certain criteria.`,
		Run: func(cmd *cobra.Command, args []string) {
			id := zfs.NewDatasetID(Pool, Dataset)
			snapKey := zfs.MakeSnapshotKey(id, Prefix)
			snapshots, err := zfs.GetSnapshotsWithKey(Pool, snapKey)
			if err != nil {
				fmt.Println(err)
				return
			}

			toDelete := pickSnapshotsToDelete(snapshots, time.Now(), saveCount, saveDuration)
			for _, snapshot := range toDelete {
				if err := zfs.DeleteSnapshot(snapshot, DryRun); err != nil {
					fmt.Println(err)
					return
				}
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(cleanupCmd)

	cleanupCmd.Flags().StringVarP(&Pool, "pool", "p", "", "name of the zfs pool")
	cleanupCmd.Flags().StringVarP(&Dataset, "dataset", "d", "", "name of the zfs dataset")
	cleanupCmd.Flags().StringVar(&Prefix, "prefix", "", "snapshot prefix")
	cleanupCmd.Flags().IntVarP(&saveCount, "count", "c", -1, "Number of snapshots to save. Defaults to saving all snapshots.")
	cleanupCmd.Flags().DurationVar(&saveDuration, "duration", 0, "Length of time to save snapshots. Snapshots older than"+
		"now - duration will be destroyed. Defaults to saving snapshots indefinitely. Overriden by count.")

	cleanupCmd.MarkFlagRequired("pool")
	cleanupCmd.MarkFlagRequired("dataset")
}

func pickSnapshotsToDelete(snapshots []string, now time.Time, count int, ageLimit time.Duration) []string {
	if len(snapshots) == 0 {
		fmt.Println("Skipping snapshot selection because no snapshots found.")
		return make([]string, 0)
	}

	if int64(saveDuration.Minutes()) > 1 && count >= 0 {
		fmt.Println("Ignoring save duration because save count specified.")
	}

	if count > 0 {
		fmt.Printf("Saving the last %d snapshots\n", count)
		sort.Sort(sort.StringSlice(snapshots))
		return snapshots[0 : len(snapshots)-count]
	} else if count == 0 {
		fmt.Println("Deleting all old snapshots.")
		return snapshots
	} else if int64(ageLimit.Minutes()) > 1 {
		fmt.Printf("Saving last %s of snapshots\n", fmtSnapshotDuration(ageLimit))
		re := regexp.MustCompile(`-(\d{14})$`)

		maxAge := now.Add(-ageLimit)
		snapsToDelete := []string{}
		for _, s := range snapshots {
			// pool/dataset@(prefix)pool-dataset-YYYYmmddHHmmss
			matches := re.FindStringSubmatch(s)
			if len(matches) < 2 || matches[1] == "" {
				log.Printf("Skipping snapshot with invalid date string: %s", s)
				continue
			}
			snapTimestamp := matches[1]

			// Parse date string into date/time
			snapDate, err := time.ParseInLocation(snapTimeFmt, snapTimestamp, time.Local)
			if err != nil {
				log.Printf("Error parsing snapshot date for snapshot %s [%s]", s, snapTimestamp)
				continue
			}

			if snapDate.Before(maxAge) {
				snapsToDelete = append(snapsToDelete, s)
			}
		}
		return snapsToDelete
	}

	fmt.Println("Saving all old snapshots")
	return make([]string, 0)
}

func fmtSnapshotDuration(d time.Duration) string {
	d2 := d.Round(time.Minute)

	weeks := d2 / (oneWeek)
	d2 -= weeks * oneWeek

	days := d2 / (oneDay)
	d2 -= days * oneDay

	hours := d2 / time.Hour
	d2 -= hours * time.Hour

	mins := d2 / time.Minute

	sb := strings.Builder{}

	if weeks > 0 {
		sb.WriteString(fmt.Sprintf("%dw", weeks))
	}
	if days > 0 {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%dh", hours))
	}
	if mins > 0 {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%dm", mins))
	}

	return sb.String()
}
