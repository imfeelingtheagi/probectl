# Example root: deploy netctl onto any Kubernetes cluster with the netctl module.
# Cloud-agnostic — point the providers at any kubeconfig (EKS/GKE/AKS/OpenShift/
# k3s). Provision the cluster + managed Postgres with your cloud's modules, then
# pass the DSN in via database_url.

terraform {
  required_version = ">= 1.5"
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig
  }
}

module "netctl" {
  source = "../../modules/netctl"

  # Local chart in this repo (so the size presets resolve).
  chart = "../../../helm/netctl"

  namespace          = "netctl"
  size               = var.size
  ingress_host       = var.ingress_host
  ingress_tls_secret = var.ingress_tls_secret

  database_url = var.database_url
  envelope_key = var.envelope_key
}

output "namespace" {
  value = module.netctl.namespace
}

output "release" {
  value = module.netctl.release_name
}

output "app_version" {
  value = module.netctl.app_version
}
