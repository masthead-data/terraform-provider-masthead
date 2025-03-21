terraform {
  required_providers {
    masthead = {
      source = "masthead-data/masthead"
      version = "0.1.0"
    }
  }
}

provider "masthead" {
  api_token = var.api_token
}

data "masthead_user" "user" {
  email        = "user@example.com"
}

resource "masthead_user" "user" {
  email        = data.masthead_user.user.email
  role         = "USER"
}

variable "api_token" {
  type        = string
  sensitive   = true
}
