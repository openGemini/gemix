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

| command    | description | parameter                                                             | example                                                                                                                           |
|------------| --- |-----------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| `install`  | install an openGemini cluster | cluser-name version topology.yml --user username<br />--password | `./gemix cluster install2 gemini-cluster v1.1.0 ./topology.example.yaml --skip-create-user --user root -p`                        |
| `start`    | start an openGemini cluster and check the running status after startup | cluster-name                            | `./gemix cluster start gemini-cluster `                                                  |
| `stop`     | stop an openGemini cluster                                   | --name                                                                | `./gemix cluster stop --name gemini-cluster`                                                                                      |
| `uninstall` | uninstall an openGemini cluster which means clearing data    | --name                                                                | `./gemix cluster uninstall --name gemini-cluster`                                                                                 |
| `status`   | Check the running status of the openGemini cluster, including port occupancy, disk capacity, etc | --name                                                                | `./gemix cluster status --name gemini-cluster`                                                                                    |
| `upgrade`  | upgrade an openGemini cluster to the specified version and uninstall the old one | --name<br />--version                                                 | `./gemix cluster upgrade --name gemini-cluster --version v1.0.1`                                                                  |

## topology.example.yaml

The `topology.example.yaml` is written by the user and contains the necessary information for deploying the openGemini cluster. You can modify the content of the file according to the template.

The meaning of each part is as follows:

* `global`: Default values for some options. These options are mandatory.
* `ts-meta`:  Deployment information for `ts-meta`, users can modify some options in `openGemini.conf` here.
* `ts-sql`: Deployment information for `ts-sql`, users can modify some options in `openGemini.conf` here.
* `ts-store`: Deployment information for `ts-store`, users can modify some options in `openGemini.conf` here.

[Go to see the example](./embed/examples/cluster/topology.example.yaml)