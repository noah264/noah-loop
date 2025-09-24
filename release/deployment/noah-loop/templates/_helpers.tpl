{{/*
Expand the name of the chart.
*/}}
{{- define "noah-loop.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "noah-loop.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "noah-loop.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "noah-loop.labels" -}}
helm.sh/chart: {{ include "noah-loop.chart" . }}
{{ include "noah-loop.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "noah-loop.selectorLabels" -}}
app.kubernetes.io/name: {{ include "noah-loop.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service selector labels for specific service
*/}}
{{- define "noah-loop.serviceSelectorLabels" -}}
app.kubernetes.io/name: {{ include "noah-loop.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "noah-loop.serviceAccountName" -}}
{{- if .Values.global.serviceAccount.create }}
{{- default (include "noah-loop.fullname" .) .Values.global.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.global.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate service name for specific component
*/}}
{{- define "noah-loop.componentName" -}}
{{- printf "%s-%s" (include "noah-loop.fullname" .) .component }}
{{- end }}

{{/*
Generate image reference
*/}}
{{- define "noah-loop.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry .repository .tag }}
{{- else }}
{{- printf "%s:%s" .repository .tag }}
{{- end }}
{{- end }}

{{/*
Generate environment variables for database connection
*/}}
{{- define "noah-loop.databaseEnvVars" -}}
- name: DATABASE_HOST
  value: {{ .Values.config.database.host | quote }}
- name: DATABASE_PORT
  value: {{ .Values.config.database.port | quote }}
- name: DATABASE_DATABASE
  value: {{ .Values.config.database.database | quote }}
- name: DATABASE_SSLMODE
  value: {{ .Values.config.database.sslmode | quote }}
- name: DATABASE_MAX_OPEN_CONNS
  value: {{ .Values.config.database.maxOpenConns | quote }}
- name: DATABASE_MAX_IDLE_CONNS
  value: {{ .Values.config.database.maxIdleConns | quote }}
- name: DATABASE_MAX_LIFETIME
  value: {{ .Values.config.database.maxLifetime | quote }}
- name: DATABASE_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "noah-loop.databaseSecretName" . }}
      key: {{ .Values.secrets.database.existingSecretUsernameKey }}
- name: DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "noah-loop.databaseSecretName" . }}
      key: {{ .Values.secrets.database.existingSecretPasswordKey }}
{{- end }}

{{/*
Generate environment variables for Redis connection
*/}}
{{- define "noah-loop.redisEnvVars" -}}
- name: REDIS_ADDR
  value: {{ .Values.config.redis.addr | quote }}
- name: REDIS_DB
  value: {{ .Values.config.redis.db | quote }}
- name: REDIS_POOL_SIZE
  value: {{ .Values.config.redis.poolSize | quote }}
- name: REDIS_MIN_IDLE_CONNS
  value: {{ .Values.config.redis.minIdleConns | quote }}
- name: REDIS_DIAL_TIMEOUT
  value: {{ .Values.config.redis.dialTimeout | quote }}
- name: REDIS_READ_TIMEOUT
  value: {{ .Values.config.redis.readTimeout | quote }}
- name: REDIS_WRITE_TIMEOUT
  value: {{ .Values.config.redis.writeTimeout | quote }}
{{- if .Values.secrets.redis.password }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "noah-loop.redisSecretName" . }}
      key: {{ .Values.secrets.redis.existingSecretPasswordKey }}
{{- end }}
{{- end }}

{{/*
Generate environment variables for etcd connection
*/}}
{{- define "noah-loop.etcdEnvVars" -}}
- name: ETCD_ENDPOINTS
  value: {{ join "," .Values.config.etcd.endpoints | quote }}
- name: ETCD_DIAL_TIMEOUT
  value: {{ .Values.config.etcd.dialTimeout | quote }}
- name: ETCD_DIAL_KEEPALIVE_TIME
  value: {{ .Values.config.etcd.dialKeepaliveTime | quote }}
- name: ETCD_DIAL_KEEPALIVE_TIMEOUT
  value: {{ .Values.config.etcd.dialKeepaliveTimeout | quote }}
- name: ETCD_MAX_CALL_SEND_MSG_SIZE
  value: {{ .Values.config.etcd.maxCallSendMsgSize | quote }}
- name: ETCD_MAX_CALL_RECV_MSG_SIZE
  value: {{ .Values.config.etcd.maxCallRecvMsgSize | quote }}
- name: ETCD_LOG_LEVEL
  value: {{ .Values.config.etcd.logLevel | quote }}
{{- end }}

{{/*
Generate environment variables for tracing
*/}}
{{- define "noah-loop.tracingEnvVars" -}}
- name: TRACING_ENABLED
  value: {{ .Values.config.tracing.enabled | quote }}
- name: TRACING_JAEGER_ENDPOINT
  value: {{ .Values.config.tracing.jaegerEndpoint | quote }}
- name: TRACING_SAMPLE_RATE
  value: {{ .Values.config.tracing.sampleRate | quote }}
{{- end }}

{{/*
Common environment variables
*/}}
{{- define "noah-loop.commonEnvVars" -}}
- name: APP_NAME
  value: {{ .Values.app.name | quote }}
- name: APP_VERSION
  value: {{ .Values.app.version | quote }}
- name: APP_ENVIRONMENT
  value: {{ .Values.app.environment | quote }}
- name: APP_DEBUG
  value: {{ .Values.app.debug | quote }}
- name: LOG_LEVEL
  value: "info"
- name: LOG_FORMAT
  value: "json"
{{- include "noah-loop.databaseEnvVars" . }}
{{- include "noah-loop.redisEnvVars" . }}
{{- include "noah-loop.etcdEnvVars" . }}
{{- include "noah-loop.tracingEnvVars" . }}
{{- end }}

{{/*
Database secret name
*/}}
{{- define "noah-loop.databaseSecretName" -}}
{{- if .Values.secrets.database.existingSecret }}
{{- .Values.secrets.database.existingSecret }}
{{- else }}
{{- printf "%s-database" (include "noah-loop.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Redis secret name
*/}}
{{- define "noah-loop.redisSecretName" -}}
{{- if .Values.secrets.redis.existingSecret }}
{{- .Values.secrets.redis.existingSecret }}
{{- else }}
{{- printf "%s-redis" (include "noah-loop.fullname" .) }}
{{- end }}
{{- end }}

{{/*
App secrets name
*/}}
{{- define "noah-loop.appSecretName" -}}
{{- if .Values.secrets.app.existingSecret }}
{{- .Values.secrets.app.existingSecret }}
{{- else }}
{{- printf "%s-app" (include "noah-loop.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Generate resource limits and requests
*/}}
{{- define "noah-loop.resources" -}}
{{- if .resources }}
resources:
  {{- if .resources.limits }}
  limits:
    {{- if .resources.limits.cpu }}
    cpu: {{ .resources.limits.cpu }}
    {{- end }}
    {{- if .resources.limits.memory }}
    memory: {{ .resources.limits.memory }}
    {{- end }}
  {{- end }}
  {{- if .resources.requests }}
  requests:
    {{- if .resources.requests.cpu }}
    cpu: {{ .resources.requests.cpu }}
    {{- end }}
    {{- if .resources.requests.memory }}
    memory: {{ .resources.requests.memory }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}
