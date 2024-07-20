terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }
}

provider "docker" {}

resource "docker_image" "postgres" {
  name         = "postgres:15.7"
  keep_locally = false
}

resource "docker_image" "postgres_exporter" {
  name         = "quay.io/prometheuscommunity/postgres-exporter"
  keep_locally = false
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_password" {
  description = "Database password"
  type        = string
}

variable "db_user" {
  description = "Database user"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = number
}

variable "db_container_name" {
  description = "Database container name"
  type        = string
}

variable "exporter_port" {
  description = "Exporter port"
  type        = number
}

variable "exporter_container_name" {
  description = "Exporter container name"
  type        = string
}

variable "node_ip" {
  description = "Node ip"
  type        = string
}

resource "docker_container" "example_db" {
  name  = var.db_container_name
  image = docker_image.postgres.image_id

  env = [
    "POSTGRES_DB=${var.db_name}",
    "POSTGRES_USER=${var.db_user}",
    "POSTGRES_PASSWORD=${var.db_password}"
  ]

  ports {
    internal = 5432
    external = var.db_port
  }
}

resource "docker_container" "postgres_exporter" {
  name  = var.exporter_container_name
  image = docker_image.postgres_exporter.image_id

  env = [
    "DATA_SOURCE_NAME=postgresql://${var.db_user}:${var.db_password}@${var.node_ip}:${var.db_port}/${var.db_name}?sslmode=disable",
  ]

  ports {
    internal = 9187
    external = var.exporter_port
  }

  depends_on = [docker_container.example_db]
}
