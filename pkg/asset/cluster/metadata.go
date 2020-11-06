package cluster

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/cluster/azure"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/types"
	azuretypes "github.com/openshift/installer/pkg/types/azure"
)

const (
	metadataFileName = "metadata.json"
)

// Metadata contains information needed to destroy clusters.
type Metadata struct {
	File *asset.File
}

var _ asset.WritableAsset = (*Metadata)(nil)

// Name returns the human-friendly name of the asset.
func (m *Metadata) Name() string {
	return "Metadata"
}

// Dependencies returns the direct dependencies for the metadata
// asset.
func (m *Metadata) Dependencies() []asset.Asset {
	return []asset.Asset{
		&installconfig.ClusterID{},
		&installconfig.InstallConfig{},
	}
}

// Generate generates the metadata asset.
func (m *Metadata) Generate(parents asset.Parents) (err error) {
	clusterID := &installconfig.ClusterID{}
	installConfig := &installconfig.InstallConfig{}
	parents.Get(clusterID, installConfig)

	metadata := &types.ClusterMetadata{
		ClusterName: installConfig.Config.ObjectMeta.Name,
		ClusterID:   clusterID.UUID,
		InfraID:     clusterID.InfraID,
	}

	switch installConfig.Config.Platform.Name() {
	case azuretypes.Name:
		metadata.ClusterPlatformMetadata.Azure = azure.Metadata(installConfig.Config)

	default:
		return errors.Errorf("no known platform")
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, "failed to Marshal ClusterMetadata")
	}

	m.File = &asset.File{
		Filename: metadataFileName,
		Data:     data,
	}

	return nil
}

// Files returns the metadata file generated by the asset.
func (m *Metadata) Files() []*asset.File {
	if m.File != nil {
		return []*asset.File{m.File}
	}
	return []*asset.File{}
}

// Load is a no-op, because we never want to load broken metadata from
// the disk.
func (m *Metadata) Load(f asset.FileFetcher) (found bool, err error) {
	return false, nil
}

// LoadMetadata loads the cluster metadata from an asset directory.
func LoadMetadata(dir string) (*types.ClusterMetadata, error) {
	path := filepath.Join(dir, metadataFileName)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata *types.ClusterMetadata
	if err = json.Unmarshal(raw, &metadata); err != nil {
		return nil, errors.Wrapf(err, "failed to Unmarshal data from %q to types.ClusterMetadata", path)
	}

	return metadata, err
}
