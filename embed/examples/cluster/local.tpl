# For more information about the format of the openGemini cluster topology file, consult
# https://docs.opengemini.org/guide/deploy_cluster/production_deployment_using_gemix.html

# # Global variables are applied to all deployments and used as the default value of
# # the deployments if a specific deployment value is missing.
global:
  # # The OS user who runs the openGemini cluster.
  user: "{{ .GlobalUser }}"
  {{- if .GlobalGroup }}
  # group is used to specify the group name the user belong to if it's not the same as user.
  group: "{{ .GlobalGroup }}"
  {{- end }}
  # # SSH port of servers in the managed cluster.
  ssh_port: {{ .GlobalSSHPort }}
  # # Storage directory for cluster deployment files, startup scripts, and configuration files.
  deploy_dir: "{{ .GlobalDeployDir }}"
  # # Log directory for cluster components.
  log_dir: "{{ .GlobalLogDir }}"
  # # openGemini Cluster data storage directory
  data_dir: "{{ .GlobalDataDir }}"
  {{- if .GlobalArch }}·
  # # Supported values: "amd64", "arm64" (default: "amd64")
  arch: "{{ .GlobalArch }}"
  {{- end }}

{{ if .TSMetaServers -}}
ts_meta_servers:
{{- range .TSMetaServers }}
  - host: {{ . }}
{{- end }}
{{ end }}
{{ if .TSSqlServers -}}
ts_sql_servers:
{{- range .TSSqlServers }}
  - host: {{ . }}
{{- end }}
{{ end }}
{{ if .TSStoreServers -}}
ts_store_servers:
{{- range .TSStoreServers }}
  - host: {{ . }}
{{- end }}
{{ end }}
{{ if .GrafanaServers -}}
grafana_servers:
 {{- range .GrafanaServers }}
  - host: {{ . }}
{{- end }}
{{ end }}