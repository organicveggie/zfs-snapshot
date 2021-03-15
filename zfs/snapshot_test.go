package zfs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMakeSnapshotKey(t *testing.T) {
	tests := []struct {
		name    string
		dataset DatasetID
		prefix  string
		want    string
	}{
		{
			name: "PoolDataset",
			dataset: DatasetID{
				pool:    "pool",
				dataset: "dataset",
			},
			want: "pool/dataset@pool-dataset",
		},
		{
			name: "PoolDatasetPrefix",
			dataset: DatasetID{
				pool:    "pool",
				dataset: "dataset",
			},
			prefix: "prefix-",
			want:   "pool/dataset@prefix-pool-dataset",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MakeSnapshotKey(test.dataset, test.prefix)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("difference -want, +got: %s", diff)
			}
		})
	}
}
