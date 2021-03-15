package zfs

// DatasetID uniquely identifies a ZFS dataset within a pool.
type DatasetID struct {
	pool    string
	dataset string
}

// NewDatasetID creates a new instance of a DatasetID.
func NewDatasetID(pool, dataset string) DatasetID {
	return DatasetID{
		pool:    pool,
		dataset: dataset,
	}
}

// GetPool returns the ZFS pool name.
func (d DatasetID) GetPool() string {
	return d.pool
}

// GetDataset returns the ZFS dataset name.
func (d DatasetID) GetDataset() string {
	return d.dataset
}
