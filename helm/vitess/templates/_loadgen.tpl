###################################
# loadgen Service + Deployment
###################################
{{- define "loadgen" -}}
# set tuple values to more recognizable variables
{{- $defaultLoadgen := index . 0 -}}

###################################
# loadgen Service
###################################
kind: Service
apiVersion: v1
metadata:
  name: loadgen
  labels:
    component: loadgen
    app: vitess
spec:
  ports:
    - name: web
      port: 24001
  selector:
    component: loadgen
    app: vitess
  type: {{$defaultLoadgen.serviceType}}
---
###################################
# loadgen Service + Deployment
###################################
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: loadgen
spec:
  replicas: {{$defaultLoadgen.replicas}}
  selector:
    matchLabels:
      app: vitess
      component: loadgen
  template:
    metadata:
      labels:
        app: vitess
        component: loadgen
    spec:
{{ include "pod-security" . | indent 6 }}
      containers:
        - name: loadgen
          image: sougou/loadgen:latest

          resources:
{{ toYaml ($defaultLoadgen.resources) | indent 12 }}
          command:
            - bash
            - "-c"
            - |
              set -ex;
              eval exec /vt/bin/loadgen

{{- end -}}
