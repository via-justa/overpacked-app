{{- define "overpacked-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "overpacked-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.labels" -}}
helm.sh/chart: {{ include "overpacked-app.chart" . }}
app.kubernetes.io/name: {{ include "overpacked-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "overpacked-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "overpacked-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "overpacked-app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "overpacked-app.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "overpacked-app.backend.fullname" -}}
{{- printf "%s-backend" (include "overpacked-app.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.frontend.fullname" -}}
{{- printf "%s-frontend" (include "overpacked-app.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.db.clusterName" -}}
{{- printf "%s-db" (include "overpacked-app.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.db.rwService" -}}
{{- printf "%s-rw" (include "overpacked-app.db.clusterName" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "overpacked-app.backend.upstream" -}}
{{- if .Values.frontend.backendUpstream -}}
{{- .Values.frontend.backendUpstream -}}
{{- else -}}
{{- printf "%s:%d" (include "overpacked-app.backend.fullname" .) (int .Values.backend.service.port) -}}
{{- end -}}
{{- end -}}

{{- define "overpacked-app.externalDatabaseUrl" -}}
{{- $sslmode := default "disable" .Values.database.external.sslmode -}}
{{- $username := required "database.auth.username is required when database.mode=external" .Values.database.auth.username | urlquery -}}
{{- $password := required "database.auth.password is required when database.mode=external" .Values.database.auth.password | urlquery -}}
{{- $host := required "database.external.host is required when database.mode=external" .Values.database.external.host -}}
{{- $port := int (default 5432 .Values.database.external.port) -}}
{{- $database := required "database.auth.database is required when database.mode=external" .Values.database.auth.database | urlquery -}}
{{- printf "postgres://%s:%s@%s:%d/%s?sslmode=%s" $username $password $host $port $database $sslmode -}}
{{- end -}}
