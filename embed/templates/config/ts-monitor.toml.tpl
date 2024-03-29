[monitor]
  # localhost ip
  host = "{{.Host}}"
  # Indicates the path of the metric file generated by the kernel. References openGemini.conf: [monitor] store-path
  # metric-path cannot have subdirectories
  metric-path = "{{.MetricPath}}"
  # Indicates the path of the log file generated by the kernel. References openGemini.conf: [logging] path
  # error-log-path cannot have subdirectories
  error-log-path = "{{.ErrorLogPath}}"
  # Data disk path. References openGemini.conf: [data] store-data-dir
  disk-path = "{{.DataPath}}"
  # Wal disk path. References openGemini.conf: [data]  store-wal-dir
  aux-disk-path = "{{.WALPath}}"
  # Name of the process to be monitored. Optional Value: ts-store,ts-sql,ts-meta.
  # Determined based on the actual process running on the local node.
  process = "{{.ProcessName}}"
  # the history file reported error-log names.
  history-file = "history.txt"
  # Is the metric compressed.
  compress = false

[query]
  # query for some DDL. Report for these data to monitor cluster.
  # - SHOW DATABASES
  # - SHOW MEASUREMENTS
  # - SHOW SERIES CARDINALITY FROM mst
  query-enable = false
  http-endpoint = "127.0.0.x:8086"
  query-interval = "5m"
  # username = ""
  # password = ""
  # https-enable = false

[report]
  # Address for metric data to be reported.
  address = "{{.MonitorAddr}}"
  # Database name for metric data to be reported.
  database = "{{.MonitorDB}}"
  rp = "autogen"
  rp-duration = "168h"
  # username = ""
  # password = ""
{{- if .TLSEnabled}}
  https-enable = true
{{- end}}

[logging]
  format = "auto"
  level = "info"
  path = "{{.LoggingPath}}"
  max-size = "64m"
  max-num = 30
  max-age = 7
  compress-enabled = true
