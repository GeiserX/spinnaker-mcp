{{/*
Expand the name of the chart.
*/}}
{{- define "spinnaker-mcp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "spinnaker-mcp.fullname" -}}
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
{{- define "spinnaker-mcp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "spinnaker-mcp.labels" -}}
helm.sh/chart: {{ include "spinnaker-mcp.chart" . }}
{{ include "spinnaker-mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "spinnaker-mcp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "spinnaker-mcp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "spinnaker-mcp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "spinnaker-mcp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image tag — defaults to v + Chart.AppVersion; custom tags used as-is.
*/}}
{{- define "spinnaker-mcp.imageTag" -}}
{{- if .Values.image.tag -}}
{{- .Values.image.tag -}}
{{- else -}}
v{{- .Chart.AppVersion -}}
{{- end -}}
{{- end }}

{{/*
The chart serves /mcp on the pod network with the Gate credential, so it refuses to
render without a bearer token for it.
*/}}
{{- define "spinnaker-mcp.requireMcpAuth" -}}
{{- if not (or .Values.mcpAuth.existingSecret .Values.mcpAuth.token) }}
{{- fail "mcpAuth.existingSecret or mcpAuth.token is required: /mcp acts with the Gate credential, so every request must carry a bearer token. Create one with: openssl rand -hex 32" }}
{{- end }}
{{- end }}

{{- define "spinnaker-mcp.mcpAuthSecretName" -}}
{{- if .Values.mcpAuth.existingSecret }}
{{- .Values.mcpAuth.existingSecret }}
{{- else }}
{{- printf "%s-mcp-auth" (include "spinnaker-mcp.fullname" .) }}
{{- end }}
{{- end }}

{{- define "spinnaker-mcp.mcpAuthSecretKey" -}}
{{- if .Values.mcpAuth.existingSecret }}
{{- .Values.mcpAuth.existingSecretKey | default "MCP_AUTH_TOKEN" }}
{{- else }}
{{- "MCP_AUTH_TOKEN" }}
{{- end }}
{{- end }}
