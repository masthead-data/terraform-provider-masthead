terraform {
  required_providers {
    masthead = {
      source = "masthead-data/masthead"
      version = "0.1.0"
    }
  }
}

variable "api_token" {
  type        = string
  sensitive   = true
}

provider "masthead" {
  api_token = var.api_token
}

resource "masthead_user" "user2" {
  email        = "user2@example.com"
  role         = "OWNER"
}
