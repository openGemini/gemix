# openGemini-UP

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
| `version` | display version number of openGemini-UP | no para | `./openGemini-UP version` |
| `list` | display the version information of all components currently downloaded | no para | `./openGemini-UP list` |
| `install` | install database components | --version | `./openGemini-UP install --version v1.0.0` |
| `cluster` | deploying and managing openGemini clusters | have subcommand | |

The following table describes the subcommands of the `cluster` command.

| command | description | parameter | example |
| --- | --- | --- | --- |
| `deploy` | deploy an openGemini cluster| --version | `./openGemini-UP cluster deploy --version v1.0.0` |
| `stop` | stop an openGemini cluster | no para | `./openGemini-UP cluster stop` |
| `start` | start an openGemini cluster which is stopped | no para | `./openGemini-UP cluster start` |
| `destroy` | destroy an openGemini cluster which means stopping services and clearing data| no para | `./openGemini-UP cluster destroy`  |
| `upgrade` | upgrade an openGemini cluster to the specified version | --version | `./openGemini-UP cluster upgrade --version v1.0.0`  |

## up.conf

The `up.conf` is written by the user and contains the necessary information for deploying the openGemini cluster. You can modify the content of the file according to the template, but please do not change the location of the file.

The meaning of each part is as follows:

* `common`: Describe the deployment distribution of the cluster on each machine.
* `host`: Describe the IP address of each machine.
* `ssh`: Describe the SSH configuration for each machin. The login methods for different machines should be the same.

```toml
[common]
  meta = ["host1","host2","host3"]
  store = ["host1","host2","host3"]
  sql = ["host1"]

[host]
  name = ["host1","host2","host3"]
  ip = ["xxx.xxx.xxx.xx1","xxx.xxx.xxx.xx2","xxx.xxx.xxx.xx3"]

[ssh]
  port = 22
  user = "username"
  type = "SSH_KEY"  # "SSH_PW"
  password = "xxxxx"
  key-path = "~/.ssh/id_rsa"
  up-data-path = "~/openGemini-UP/"
```
