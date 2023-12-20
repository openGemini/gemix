apiVersion: 1
datasources:
  - name: {{.ClusterName}}
    type: influxdb
    access: proxy
    url: {{.URL}}
    database: {{.ClusterName}}
    jsonData:
      dbName: {{.ClusterName}}
      httpMode: GET
    withCredentials: false
    isDefault: false
    tlsAuth: false
    tlsAuthWithCACert: false
    version: 1
    editable: true