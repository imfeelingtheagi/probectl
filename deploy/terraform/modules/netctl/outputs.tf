output "namespace" {
  description = "Namespace the release is deployed in."
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name."
  value       = helm_release.netctl.name
}

output "release_status" {
  description = "Helm release status."
  value       = helm_release.netctl.status
}

output "chart_version" {
  description = "Resolved chart version."
  value       = helm_release.netctl.metadata[0].version
}

output "app_version" {
  description = "Deployed app (netctl) version."
  value       = helm_release.netctl.metadata[0].app_version
}

output "secret_name" {
  description = "Name of the Kubernetes Secret holding the sensitive config."
  value       = kubernetes_secret.netctl.metadata[0].name
}
