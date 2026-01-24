{{/*
Expand the name of the chart.
*/}}
{{- define "boilerr.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "boilerr.fullname" -}}
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
{{- define "boilerr.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "boilerr.labels" -}}
helm.sh/chart: {{ include "boilerr.chart" . }}
{{ include "boilerr.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: boilerr
app.kubernetes.io/component: operator
{{- end }}

{{/*
Selector labels
*/}}
{{- define "boilerr.selectorLabels" -}}
app.kubernetes.io/name: {{ include "boilerr.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "boilerr.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "boilerr.fullname" . | printf "%s-controller-manager") .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace
*/}}
{{- define "boilerr.namespace" -}}
{{- if .Values.namespaceOverride }}
{{- .Values.namespaceOverride }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Image reference
*/}}
{{- define "boilerr.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Check if a game is enabled
*/}}
{{- define "boilerr.isGameEnabled" -}}
{{- $game := .game -}}
{{- $enabled := false -}}
{{- if .root.Values.gameDefinitions.enabled }}
{{- $allGames := list "valheim" }}
{{- if has $game $allGames }}
{{- if gt (len .root.Values.gameDefinitions.include) 0 }}
{{- if has $game .root.Values.gameDefinitions.include }}
{{- $enabled = true }}
{{- end }}
{{- else }}
{{- $enabled = true }}
{{- end }}
{{- if has $game .root.Values.gameDefinitions.exclude }}
{{- $enabled = false }}
{{- end }}
{{- end }}
{{- end }}
{{- $enabled }}
{{- end }}
