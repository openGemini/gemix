package config

import (
	"openGemini-UP/util"
	"path/filepath"
	"strconv"
	"strings"
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

func GenConfs(y Yaml, template Toml, path string) {
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

	// generate corresponding openGemini.conf for ervery host.
	for _, host := range hosts {
		fileName := filepath.Join(path, host.Ip+util.Remote_conf_suffix)
		newToml := template
		addr := host.Ip

		// common
		newToml.Common.MetaJoin = metaJoins
		// meta
		if host.HasMeta {
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

			if host.TsMeta.LogDir != "" {
				newToml.Logging.Path = host.TsMeta.LogDir
			}

			if host.TsMeta.GossipPort != 0 {
				newToml.Gossip.MetaBindPort = host.TsMeta.GossipPort
			}
		}

		// http
		if host.HasSql {
			if host.TsSql.Port != 0 {
				newToml.Http.BindAddress = addr + ":" + strconv.Itoa(host.TsSql.Port)
			} else {
				newToml.Http.BindAddress = strings.Replace(newToml.Http.BindAddress, newToml.Http.BindAddress[:8], addr, -1)
			}

			if host.TsSql.LogDir != "" {
				newToml.Logging.Path = host.TsSql.LogDir
			}
		}

		// data
		if host.HasStore {
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

			if host.TsStore.LogDir != "" {
				newToml.Logging.Path = host.TsStore.LogDir
			}

			if host.TsStore.GossipPort != 0 {
				newToml.Gossip.StoreBindPort = host.TsStore.GossipPort
			}
		}

		// logging (already processed)

		// gossip
		newToml.Gossip.BindAddress = addr
		newToml.Gossip.Members = gossipMembers

		GenNewToml(newToml, fileName)
	}
}
