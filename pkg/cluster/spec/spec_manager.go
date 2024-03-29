// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	errNS = errorx.NewNamespace("spec")
	// ErrCreateDirFailed is ErrCreateDirFailed
	ErrCreateDirFailed = errNS.NewType("create_dir_failed")
	// ErrSaveMetaFailed is ErrSaveMetaFailed
	ErrSaveMetaFailed = errNS.NewType("save_meta_failed")
	// ErrSaveScaleOutFileFailed is ErrSaveMetaFailed
	ErrSaveScaleOutFileFailed = errNS.NewType("save_scale-out_lock_failed")
)

const (
	// metaFileName is the file name of the meta file.
	metaFileName = "meta.yaml"
	// BackupDirName is the directory to save backup files.
	BackupDirName = "backup"
)

// SpecManager control management of spec meta data.
type SpecManager struct {
	base    string
	newMeta func() *ClusterMeta
}

// NewSpec create a spec instance.
func NewSpec(base string, newMeta func() *ClusterMeta) *SpecManager {
	return &SpecManager{
		base:    base,
		newMeta: newMeta,
	}
}

// NewMetadata alloc a Metadata according the type.
func (s *SpecManager) NewMetadata() *ClusterMeta {
	return s.newMeta()
}

// Path returns the full path to a sub path (file or directory) of a
// cluster, it is a sub dir in the profile dir of the user, with the cluster name
// as its name.
func (s *SpecManager) Path(cluster string, subpath ...string) string {
	if cluster == "" {
		cluster = "default-cluster"
	}
	return filepath.Join(append([]string{s.base, cluster}, subpath...)...)
}

// SaveMeta save the meta with specified cluster name.
func (s *SpecManager) SaveMeta(clusterName string, meta *ClusterMeta) error {
	wrapError := func(err error) *errorx.Error {
		return ErrSaveMetaFailed.Wrap(err, "Failed to save cluster metadata")
	}

	metaFile := s.Path(clusterName, metaFileName)
	backupDir := s.Path(clusterName, BackupDirName)

	if err := s.ensureDir(clusterName); err != nil {
		return wrapError(err)
	}

	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return wrapError(err)
	}

	data, err := yaml.Marshal(meta)
	if err != nil {
		return wrapError(err)
	}

	err = utils.SaveFileWithBackup(metaFile, data, backupDir)
	if err != nil {
		return wrapError(err)
	}
	return nil
}

// Metadata tries to read the metadata of a cluster from file
func (s *SpecManager) Metadata(clusterName string, meta any) error {
	fname := s.Path(clusterName, metaFileName)

	yamlFile, err := os.ReadFile(fname)
	if err != nil {
		return errors.WithMessagef(err, "cluster name: %s read metadata failed", clusterName)
	}

	err = yaml.Unmarshal(yamlFile, meta)
	if err != nil {
		return errors.WithMessagef(err, "ccluster name: %s unmarshal metadata failed", clusterName)
	}

	return nil
}

// Exist checks if the cluster exist by checking the meta file.
func (s *SpecManager) Exist(clusterName string) (exist bool, err error) {
	fname := s.Path(clusterName, metaFileName)

	_, err = os.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}

	return true, nil
}

// Remove removes the data with specified cluster name.
func (s *SpecManager) Remove(clusterName string) error {
	return os.RemoveAll(s.Path(clusterName))
}

// List return the cluster names.
func (s *SpecManager) List() (clusterNames []string, err error) {
	fileInfos, err := os.ReadDir(s.base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	for _, info := range fileInfos {
		if utils.IsNotExist(s.Path(info.Name(), metaFileName)) {
			continue
		}
		clusterNames = append(clusterNames, info.Name())
	}

	return
}

// GetAllClusters get a metadata list of all clusters deployed by current user
func (s *SpecManager) GetAllClusters() (map[string]Metadata, error) {
	clusters := make(map[string]Metadata)
	names, err := s.List()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, name := range names {
		metadata := s.NewMetadata()
		err = s.Metadata(name, metadata)
		// clusters with topology validation errors should also be listed
		if err != nil {
			return nil, errors.WithStack(err)
		}
		clusters[name] = metadata
	}
	return clusters, nil
}

// ensureDir ensures that the cluster directory exists.
func (s *SpecManager) ensureDir(clusterName string) error {
	if err := os.MkdirAll(s.Path(clusterName), 0750); err != nil {
		return ErrCreateDirFailed.
			Wrap(err, "Failed to create cluster metadata directory '%s'", s.Path(clusterName)).
			WithProperty(gui.SuggestionFromString("Please check file system permissions and try again."))
	}
	return nil
}

var openGeminiSpec *SpecManager

// GetSpecManager return the spec manager of openGemini cluster.
func GetSpecManager() *SpecManager {
	if !initialized {
		panic("must Initialize profile first")
	}
	return openGeminiSpec
}

// ClusterMeta is the specification of generic cluster metadata
type ClusterMeta struct {
	User     string         `yaml:"user"`               // the user to run and manage cluster on remote
	Version  string         `yaml:"openGemini_version"` // the version of openGemini cluster
	Topology *Specification `yaml:"topology"`
}

// GetTopology implement Metadata interface.
func (m *ClusterMeta) GetTopology() Topology {
	return m.Topology
}

// SetTopology implement Metadata interface.
func (m *ClusterMeta) SetTopology(topo Topology) {
	myTopo, ok := topo.(*Specification)
	if !ok {
		panic(fmt.Sprintln("wrong type: ", reflect.TypeOf(topo)))
	}

	m.Topology = myTopo
}

// GetBaseMeta implements Metadata interface.
func (m *ClusterMeta) GetBaseMeta() *BaseMeta {
	return &BaseMeta{
		Version: m.Version,
		User:    m.User,
	}
}

// SetVersion implement UpgradableMetadata interface.
func (m *ClusterMeta) SetVersion(s string) {
	m.Version = s
}

// SetUser implement UpgradableMetadata interface.
func (m *ClusterMeta) SetUser(s string) {
	m.User = s
}
