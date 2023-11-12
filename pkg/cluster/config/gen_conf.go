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

package config

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openGemini/gemix/util"
)

type HostFile struct {
	Ip       string
	HasMeta  bool
	HasSql   bool
	HasStore bool
	TsMeta   MetaYaml
	TsSql    SqlYaml
	TsStore  StoreYaml
}

type MetaPorts struct {
	Ip     string
	Peer   int
	Gossip int
}

func GenConfs(y Yaml, template Toml, path string) error {
	hosts := make(map[string]HostFile)
	var metaPorts []MetaPorts

	// Build hosts
	for _, meta := range y.TsMeta {
		hostEntry, ok := hosts[meta.Host]
		if !ok {
			hostEntry = HostFile{Ip: meta.Host}
		}
		hostEntry.TsMeta = meta
		hostEntry.HasMeta = true
		hosts[meta.Host] = hostEntry

		metaPorts = append(metaPorts, MetaPorts{Ip: meta.Host, Peer: meta.PeerPort, Gossip: meta.GossipPort})
	}
	for _, sql := range y.TsSql {
		hostEntry, ok := hosts[sql.Host]
		if !ok {
			hostEntry = HostFile{Ip: sql.Host}
		}
		hostEntry.TsSql = sql
		hostEntry.HasSql = true
		hosts[sql.Host] = hostEntry
	}
	for _, store := range y.TsStore {
		hostEntry, ok := hosts[store.Host]
		if !ok {
			hostEntry = HostFile{Ip: store.Host}
		}
		hostEntry.TsStore = store
		hostEntry.HasStore = true
		hosts[store.Host] = hostEntry
	}

	// build meta-join and gossip.member
	var metaJoins []string
	var gossipMembers []string
	for i := 0; i < len(metaPorts); i++ {
		if metaPorts[i].Peer == 0 {
			metaJoins = append(metaJoins, strings.Replace(template.Common.MetaJoin[i], template.Common.MetaJoin[i][:15], metaPorts[i].Ip, -1))
		} else {
			metaJoins = append(metaJoins, metaPorts[i].Ip+":"+strconv.Itoa(metaPorts[i].Peer))
		}
		if metaPorts[i].Gossip == 0 {
			gossipMembers = append(gossipMembers, strings.Replace(template.Gossip.Members[i], template.Gossip.Members[i][:15], metaPorts[i].Ip, -1))
		} else {
			gossipMembers = append(gossipMembers, metaPorts[i].Ip+":"+strconv.Itoa(metaPorts[i].Gossip))
		}
	}

	// generate corresponding config files for every host.
	for _, host := range hosts {
		addr := host.Ip
		// meta
		if host.HasMeta {
			newToml := template

			// meta
			if host.TsMeta.ClientPort != 0 {
				newToml.Meta.HttpBindAddress = addr + ":" + strconv.Itoa(host.TsMeta.ClientPort)
			} else {
				newToml.Meta.HttpBindAddress = strings.Replace(newToml.Meta.HttpBindAddress, newToml.Meta.HttpBindAddress[:8], addr, -1)
			}
			if host.TsMeta.PeerPort != 0 {
				newToml.Meta.RpcBindAddress = addr + ":" + strconv.Itoa(host.TsMeta.PeerPort)
			} else {
				newToml.Meta.RpcBindAddress = strings.Replace(newToml.Meta.RpcBindAddress, newToml.Meta.RpcBindAddress[:8], addr, -1)
			}
			if host.TsMeta.RaftPort != 0 {
				newToml.Meta.BindAddress = addr + ":" + strconv.Itoa(host.TsMeta.RaftPort)
			} else {
				newToml.Meta.BindAddress = strings.Replace(newToml.Meta.BindAddress, newToml.Meta.BindAddress[:8], addr, -1)
			}
			if host.TsMeta.DataDir != "" {
				newToml.Meta.Dir = host.TsMeta.DataDir
			}

			// common
			newToml.Common.MetaJoin = metaJoins
			newToml.Common.HaPolicy = host.TsMeta.Config.HaPolicy

			// logging
			if host.TsMeta.LogDir != "" {
				newToml.Logging.Path = host.TsMeta.LogDir
			}

			// gossip
			if host.TsMeta.GossipPort != 0 {
				newToml.Gossip.MetaBindPort = host.TsMeta.GossipPort
			}
			newToml.Gossip.BindAddress = addr
			newToml.Gossip.Members = gossipMembers

			// monitor
			newToml.Monitor = MonitorToml{Pushers: host.TsMeta.Config.Pushers, StoreEnabled: host.TsMeta.Config.StoreEnabled,
				StoreDatabase: host.TsMeta.Config.StoreDatabase,
				StoreInterval: host.TsMeta.Config.StoreInterval,
				StorePath:     host.TsMeta.Config.StorePath,
				Compress:      host.TsMeta.Config.Compress,
				HttpEndpoint:  host.TsMeta.Config.HttpEndpoint,
				Username:      host.TsMeta.Config.Username,
				Password:      host.TsMeta.Config.Password}

			fileName := filepath.Join(path, host.Ip, util.RemoteMetaConfName)
			if err := GenNewToml(newToml, fileName); err != nil {
				return err
			}
		}

		// http
		if host.HasSql {
			newToml := template

			// http
			if host.TsSql.Port != 0 {
				newToml.Http.BindAddress = addr + ":" + strconv.Itoa(host.TsSql.Port)
			} else {
				newToml.Http.BindAddress = strings.Replace(newToml.Http.BindAddress, newToml.Http.BindAddress[:8], addr, -1)
			}
			newToml.Http.AuthEnabled = host.TsSql.Config.AuthEnabled
			newToml.Http.HttpsEnabled = host.TsSql.Config.HttpsEnabled
			newToml.Http.HttpsCertificate = host.TsSql.Config.HttpsCertificate
			newToml.Http.HttpsPrivateKey = host.TsSql.Config.HttpsPrivateKey

			// common
			newToml.Common.MetaJoin = metaJoins
			newToml.Common.HaPolicy = host.TsSql.Config.HaPolicy

			// logging
			if host.TsSql.LogDir != "" {
				newToml.Logging.Path = host.TsSql.LogDir
			}

			// gossip
			newToml.Gossip.BindAddress = addr
			newToml.Gossip.Members = gossipMembers

			// monitor
			newToml.Monitor = MonitorToml{Pushers: host.TsSql.Config.Pushers, StoreEnabled: host.TsSql.Config.StoreEnabled,
				StoreDatabase: host.TsSql.Config.StoreDatabase,
				StoreInterval: host.TsSql.Config.StoreInterval,
				StorePath:     host.TsSql.Config.StorePath,
				Compress:      host.TsSql.Config.Compress,
				HttpEndpoint:  host.TsSql.Config.HttpEndpoint,
				Username:      host.TsSql.Config.Username,
				Password:      host.TsSql.Config.Password}

			fileName := filepath.Join(path, host.Ip, util.RemoteSqlConfName)
			if err := GenNewToml(newToml, fileName); err != nil {
				return err
			}
		}

		// data
		if host.HasStore {
			newToml := template

			// data
			if host.TsStore.IngestPort != 0 {
				newToml.Data.StoreIngestAddr = addr + ":" + strconv.Itoa(host.TsStore.IngestPort)
			} else {
				newToml.Data.StoreIngestAddr = strings.Replace(newToml.Data.StoreIngestAddr, newToml.Data.StoreIngestAddr[:8], addr, -1)
			}
			if host.TsStore.SelectPort != 0 {
				newToml.Data.StoreSelectAddr = addr + ":" + strconv.Itoa(host.TsStore.SelectPort)
			} else {
				newToml.Data.StoreSelectAddr = strings.Replace(newToml.Data.StoreSelectAddr, newToml.Data.StoreSelectAddr[:8], addr, -1)
			}
			if host.TsStore.DataDir != "" {
				newToml.Data.StoreDataDir = host.TsStore.DataDir
				newToml.Data.StoreWalDir = host.TsStore.DataDir
			}
			if host.TsStore.MetaDir != "" {
				newToml.Data.StoreMetaDir = host.TsStore.MetaDir
			}

			// common
			newToml.Common.MetaJoin = metaJoins
			newToml.Common.HaPolicy = host.TsStore.Config.HaPolicy

			// logging
			if host.TsStore.LogDir != "" {
				newToml.Logging.Path = host.TsStore.LogDir
			}

			// retention
			newToml.Retention.CheckInterval = host.TsStore.Config.CheckInterval

			// gossip
			if host.TsStore.GossipPort != 0 {
				newToml.Gossip.StoreBindPort = host.TsStore.GossipPort
			}
			newToml.Gossip.BindAddress = addr
			newToml.Gossip.Members = gossipMembers

			// monitor
			newToml.Monitor = MonitorToml{Pushers: host.TsStore.Config.Pushers, StoreEnabled: host.TsStore.Config.StoreEnabled,
				StoreDatabase: host.TsStore.Config.StoreDatabase,
				StoreInterval: host.TsStore.Config.StoreInterval,
				StorePath:     host.TsStore.Config.StorePath,
				Compress:      host.TsStore.Config.Compress,
				HttpEndpoint:  host.TsStore.Config.HttpEndpoint,
				Username:      host.TsStore.Config.Username,
				Password:      host.TsStore.Config.Password}

			fileName := filepath.Join(path, host.Ip, util.RemoteStoreConfName)
			if err := GenNewToml(newToml, fileName); err != nil {
				return err
			}
		}
	}
	return nil
}
