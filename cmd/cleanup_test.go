package cmd

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestFmtSnapshotDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "3min",
			d:    3 * time.Minute,
			want: "3m",
		},
		{
			name: "3hours",
			d:    3 * time.Hour,
			want: "3h",
		},
		{
			name: "1day",
			d:    24 * time.Hour,
			want: "1d",
		},
		{
			name: "1week",
			d:    24 * 7 * time.Hour,
			want: "1w",
		},
		{
			name: "3w2d1h4m",
			d:    24*7*3*time.Hour + 24*2*time.Hour + time.Hour + 4*time.Minute,
			want: "3w 2d 1h 4m",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := fmtSnapshotDuration(test.d)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("difference -want, +got: %s", diff)
			}
		})
	}
}

func TestPickSnapshotsToDelete(t *testing.T) {
	tests := []struct {
		name     string
		snaps    []string
		now      time.Time
		count    int
		ageLimit time.Duration
		want     []string
	}{
		{
			name: "CountZeroDeleteAll",
			snaps: []string{
				"p/d@p-d-20210311200500",
				"p/d@p-d-20210310200500",
			},
			now:      time.Date(2021, 3, 12, 8, 00, 0, 0, time.Local),
			count:    0,
			ageLimit: 0,
			want: []string{
				"p/d@p-d-20210311200500",
				"p/d@p-d-20210310200500",
			},
		},
		{
			name: "CountOneDeleteAllButOne",
			snaps: []string{
				"zp1/ds1@zp1-ds1s-20210312081617",
				"zp1/ds1@zp1-ds1s-20210312081654",
				"zp1/ds1@zp1-ds1s-20210312081825",
				"zp1/ds1@zp1-ds1s-20210312081837",
				"zp1/ds1@zp1-ds1s-20210312081925",
			},
			now:      time.Date(2021, 3, 12, 8, 00, 0, 0, time.Local),
			count:    1,
			ageLimit: 0,
			want: []string{
				"zp1/ds1@zp1-ds1s-20210312081617",
				"zp1/ds1@zp1-ds1s-20210312081654",
				"zp1/ds1@zp1-ds1s-20210312081825",
				"zp1/ds1@zp1-ds1s-20210312081837",
			},
		},
		{
			name: "CountOverridesAge",
			snaps: []string{
				"p/d@p-d-20210310200500",
				"p/d@p-d-20210311200500",
			},
			now:      time.Date(2021, 3, 12, 8, 00, 0, 0, time.Local),
			count:    1,
			ageLimit: 72 * time.Hour,
			want: []string{
				"p/d@p-d-20210310200500",
			},
		},
		{
			name: "SaveLast1h",
			snaps: []string{
				"p/d@p-d-20210301200500",
				"p/d@p-d-20210310200500",
				"p/d@p-d-20210311200500",
				"p/d@p-d-20210312030500",
				"p/d@p-d-20210312080000",
			},
			now:      time.Date(2021, 3, 12, 8, 02, 0, 0, time.Local),
			count:    -1,
			ageLimit: time.Hour,
			want: []string{
				"p/d@p-d-20210301200500",
				"p/d@p-d-20210310200500",
				"p/d@p-d-20210311200500",
				"p/d@p-d-20210312030500",
			},
		},
		{
			name: "SaveAll",
			snaps: []string{
				"p/d@p-d-20210311200500",
				"p/d@p-d-20210310200500",
			},
			now:      time.Date(2021, 3, 12, 8, 0, 0, 0, time.Local),
			count:    -1,
			ageLimit: 0,
			want:     []string{},
		},
		{
			name:     "Save1of0",
			snaps:    []string{},
			now:      time.Date(2021, 3, 15, 8, 0, 0, 0, time.Local),
			count:    1,
			ageLimit: 0,
			want:     []string{},
		},
		{
			name:     "SaveLast4Hoursof0",
			snaps:    []string{},
			now:      time.Date(2021, 3, 15, 8, 0, 0, 0, time.Local),
			count:    -1,
			ageLimit: 4 * time.Hour,
			want:     []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := pickSnapshotsToDelete(test.snaps, test.now, test.count, test.ageLimit)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("difference -want, +got: %s", diff)
			}
		})
	}
}
