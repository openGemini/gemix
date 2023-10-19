# Gemix

Cluster deployment and upgrade tool.

## Build

```bash
git clone git@github.com:openGemini/openGemini-UP.git

cd openGemini-UP

go mod tidy

go build
```

## Commands

The following table describes some commonly used basic commands.

| command | description | parameter | example |
| --- | --- | --- | --- |
| `version` | display version number of gemix | no para | `./gemix version` |
| `list` | display the available components of gemix | no para | `./gemix list` |
| `cluster` | deploying and managing openGemini clusters                   | have subcommand |                           |

The following table describes the subcommands of the `cluster` command.

| command | description | parameter | example |
| --- | --- | --- | --- |
| `install` | install an openGemini cluster | --name<br />--version<br />--yaml<br />--user<br />--key<br />--password | `./gemix cluster install --name gemini-cluster --version v1.0.0 --yaml ./topology.example.yaml --user gemini --key ~/.ssh/id_rsa` |
| `start`     | start an openGemini cluster and check the running status after startup | --name<br />--skip-create-user | `./gemix cluster start --name gemini-cluster --skip-create-user` |
| `stop`      | stop an openGemini cluster                                   | --name          | `./gemix cluster stop --name gemini-cluster` |
| `uninstall` | uninstall an openGemini cluster which means clearing data    | --name | `./gemix cluster uninstall --name gemini-cluster` |
| `status`    | Check the running status of the openGemini cluster, including port occupancy, disk capacity, etc | --name | `./gemix cluster status --name gemini-cluster` |
| `upgrade`   | upgrade an openGemini cluster to the specified version and uninstall the old one | --name<br />--version | `./gemix cluster upgrade --name gemini-cluster --version v1.0.1` |

## topology.example.yaml

The `topology.example.yaml` is written by the user and contains the necessary information for deploying the openGemini cluster. You can modify the content of the file according to the template.

The meaning of each part is as follows:

* `global`: Default values for some options. These options are mandatory.
* `ts-meta`:  Deployment information for `ts-meta`, users can modify some options in `openGemini.conf` here.
* `ts-sql`: Deployment information for `ts-sql`, users can modify some options in `openGemini.conf` here.
* `ts-store`: Deployment information for `ts-store`, users can modify some options in `openGemini.conf` here.

```yaml
# Global variables are applied to all deployments and used as the default value of
# the deployments if a specific deployment value is missing.
global:
  # SSH port of servers in the managed cluster.
  ssh_port: 22
  # user who started the openGemini process.
  user: "gemini"
  # openGemini Cluster log file storage directory.
  log_dir: "/gemini-deploy/logs"
  # Storage directory for cluster deployment files, startup scripts, and configuration files.
  deploy_dir: "/gemini-deploy"
  # operating system, linux/darwin.
  os: "linux" 
  # Supported values: "amd64", "arm64" (default: "amd64").
  arch: "amd64"

# Server configs are used to specify the configuration of ts-meta Servers.
ts-meta:
  # The ip address of the ts-meta Server.
  - host: 10.0.1.11
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.12
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.13
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [meta].http-bind-address in openGemini.conf.
    client_port: 8091
    # [meta].rpc-bind-address in openGemini.conf.
    peer_port: 8092
    # [meta].bind-address in openGemini.conf.
    raft_port: 8088
    # [gossip].meta-bind-port in openGemini.conf.
    gossip_port: 8010
    # [meta].dir in openGemini.conf.
    data_dir: "/gemini-data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"

# Server configs are used to specify the configuration of ts-sql Servers.
ts-sql:
  # The ip address of the ts-sql Server.
  - host: 10.0.1.14
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.15
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.16
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [http].bind-address in openGemini.conf.
    port: 8086
    # [http].flight-address in openGemini.conf.
    flight_port: 8087
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"

# Server configs are used to specify the configuration of ts-store Servers.
ts-store:
  # The ip address of the ts-store Server.
  - host: 10.0.1.17
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.18
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
  - host: 10.0.1.19
    # SSH port of the server. (same on same server)
    ssh_port: 22
    # [data].store-ingest-addr in openGemini.conf.
    ingest_port: 8400
    # [data].store-select-addr in openGemini.conf.
    select_port: 8401
    # [gossip].store-bind-port in openGemini.conf.
    gossip_port: 8011
    # [data].store-data-dir & [data].store-wal-dir in openGemini.conf.
    data_dir: "/gemini-data/data"
    # [data].store-meta-dir in openGemini.conf.
    meta_dir: "/gemini-data/data/meta"
    # openGemini Cluster log file storage directory. (same on same server)
    log_dir: "/gemini-deploy/log"
    # Storage directory for cluster deployment files, startup scripts, and configuration files. (same on same server)
    deploy_dir: "/gemini-deploy"
```
