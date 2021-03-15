package zfs

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GetSnapshotsWithKey retrieves a list of snapshots with a given key identifying the group of
// snapshots.
func GetSnapshotsWithKey(pool, key string) ([]string, error) {
	fmt.Printf("Searching snapshots with key %s...\n", key)

	cmd := exec.Command("zfs", "list", "-H", "-o", "name", "-t", "snapshot", "-r", pool)

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
		if !strings.Contains(snap, key) {
			continue
		}

		snapshots = append(snapshots, snap)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading standard input: %w", err)
	}
	return snapshots, nil
}

// MakeSnapshotKey builds a key identifying a group of snapshots.
func MakeSnapshotKey(id DatasetID, prefix string) string {
	snapshotStr := fmt.Sprintf("%s-%s", id.pool, id.dataset)
	if prefix != "" {
		snapshotStr = prefix + snapshotStr
	}
	return fmt.Sprintf("%s/%s@%s", id.pool, id.dataset, snapshotStr)
}

// TakeSnapshot takes a ZFS snapshot of the specified pool/dataset.
func TakeSnapshot(snapKey string, now time.Time, recursive, dryRun bool) error {
	timestamp := now.Format("20060102030405")
	newSnapName := fmt.Sprintf("%s-%s", snapKey, timestamp)

	recursiveSnap := ""
	if recursive {
		recursiveSnap = "-r"
	}

	if dryRun {
		fmt.Print("[Dry Run] ")
	}
	fmt.Printf("zfs snapshot %s %s\n", recursiveSnap, newSnapName)

	if !dryRun {
		cmd := exec.Command("zfs", "snapshot", recursiveSnap, newSnapName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error executing zfs snapshot %s: %v", newSnapName, err)
		}
	}

	return nil
}

// DeleteSnapshot destroys an existing ZFS snapshot. If dryRun is enabled, does not actually
// destroy the snapshot; simply prints the command and returns.
func DeleteSnapshot(snapshot string, dryRun bool) error {
	if dryRun {
		fmt.Print("[Dry Run] ")
	}
	fmt.Printf("$ zfs destroy %s\n", snapshot)

	cmd := exec.Command("zfs", "destroy", snapshot)
	if !dryRun {
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error deleting snapshots %s: %w", snapshot, err)
		}
	}

	return nil
}
