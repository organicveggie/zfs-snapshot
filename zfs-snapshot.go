package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	oneDay  = 24 * time.Hour
	oneWeek = 24 * 7 * time.Hour

	snapTimeFmt = "20060102150405"
)

var (
	pool      = flag.String("pool", "", "Name of the ZFS pool")
	dataset   = flag.String("dataset", "", "Name of the ZFS dataset")
	prefix    = flag.String("prefix", "", "Prefix to use when crafting the name of the snapshot")
	recursive = flag.Bool("recursive", true, "Enable recursive snapshots")

	saveCount = flag.Int("count", -1, "Number of snapshots to save. Defaults to saving all snapshots. "+
		"Overrides duration.")
	saveDuration = flag.Duration("duration", 0, "Length of time to save snapshots. Snapshots older than "+
		" now - duration will be destroyed. Defaults to saving snapshots indefinitely. Overriden by count.")

	snapshotEnabled = flag.Bool("snapshot", true, "Take ZFS snapshot of the specified pool and dataset")
	cleanupEnabled  = flag.Bool("cleanup", true, "Remove older snapshots")

	dryRun = flag.Bool("dryrun", false, "Show the steps without executing them")
)

type zfs struct {
	dryRun    bool
	snapshot  bool
	recursive bool
	cleanup   bool

	pool    string
	dataset string
}

// Returns the snapshot key, without timestamp.
func (z *zfs) takeSnapshot(prefix string, now time.Time) (string, error) {
	if !z.snapshot {
		fmt.Println("Snapshot disabled via flag. Skipping snapshot.")
		return "", nil
	}

	snapshotStr := fmt.Sprintf("%s-%s", z.pool, z.dataset)
	if prefix != "" {
		snapshotStr = prefix + snapshotStr
	}
	snapKey := fmt.Sprintf("%s/%s@%ss", z.pool, z.dataset, snapshotStr)

	timestamp := now.Format("20060102030405")
	newSnapName := fmt.Sprintf("%s-%s", snapKey, timestamp)

	recursiveSnap := ""
	if z.recursive {
		recursiveSnap = "-r"
	}

	if z.dryRun {
		fmt.Print("(Dry Run) ")
	}
	fmt.Printf("zfs snapshot %s %s\n", recursiveSnap, newSnapName)

	if !z.dryRun {
		cmd := exec.Command("zfs", "snapshot", recursiveSnap, newSnapName)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("error executing zfs snapshot %s: %v", newSnapName, err)
		}
	}

	return snapKey, nil
}

func (z *zfs) getSnapshots(snapKey string) ([]string, error) {
	cmd := exec.Command("zfs", "list", "-H", "-o", "name", "-t", "snapshot", "-r", z.pool)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error connecting to standard out: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("unable execute zfs: %w", err)
	}

	snapshots := []string{}
	scanner := bufio.NewScanner(stdout)

	// Iterate over results
	for scanner.Scan() {
		snap := scanner.Text()

		// Skip pool snapshots that don't include the snapshot key
		if !strings.Contains(snap, snapKey) {
			continue
		}

		snapshots = append(snapshots, snap)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading standard input: %w", err)
	}
	return snapshots, nil
}

func (z *zfs) deleteSnapshot(snap string) error {
	if !z.cleanup {
		fmt.Println("Snapshot cleanup disabled by flag.")
		return nil
	}

	if z.dryRun {
		fmt.Print("(Dry Run) ")
	}
	fmt.Printf("$ zfs destroy %s\n", snap)

	cmd := exec.Command("zfs", "destroy", snap)
	if !z.dryRun && z.cleanup {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error deleting snapshots %s: %w", snap, err)
		}
	}

	return nil
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

func pickSnapshotsToDelete(snaps []string, now time.Time, count int, ageLimit time.Duration) []string {
	if int64(saveDuration.Minutes()) > 1 && *saveCount >= 0 {
		fmt.Println("Ignoring save duration because save_countspecified.")
	}

	if count > 0 {
		fmt.Printf("Saving the last %d snapshots\n", count)
		sort.Sort(sort.StringSlice(snaps))
		return snaps[0 : len(snaps)-count]
	} else if count == 0 {
		fmt.Println("Deleting all old snapshots.")
		return snaps
	} else if int64(ageLimit.Minutes()) > 1 {
		fmt.Printf("Saving last %s of snapshots\n", fmtSnapshotDuration(ageLimit))
		re := regexp.MustCompile(`-(\d{14})$`)

		maxAge := now.Add(-ageLimit)
		snapsToDelete := []string{}
		for _, s := range snaps {
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

func main() {
	flag.Parse()

	if *pool == "" || *dataset == "" {
		flag.Usage()
		return
	}

	now := time.Now()
	if *dryRun {
		fmt.Println("Dry Run enabled. Not making changes.")
	}

	z := &zfs{
		dryRun:    *dryRun,
		snapshot:  *snapshotEnabled,
		recursive: *recursive,
		cleanup:   *cleanupEnabled,
		pool:      *pool,
		dataset:   *dataset,
	}

	// Take snapshot
	snapKey, err := z.takeSnapshot(*prefix, now)
	if err != nil {
		fmt.Printf("Error taking snapshot: %v", err)
	}

	// Retrieve old snapshots
	oldSnapshots, err := z.getSnapshots(snapKey)
	if err != nil {
		fmt.Printf("Error retrieving snapshots: %v", err)
		return
	}
	fmt.Println("Found:")
	for _, s := range oldSnapshots {
		fmt.Printf("  - %s\n", s)
	}

	toDelete := pickSnapshotsToDelete(oldSnapshots, now, *saveCount, *saveDuration)
	for _, s := range toDelete {
		z.deleteSnapshot(s)
	}
}
