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

resource "masthead_user" "example_user" {
  email        = "user@example.com"
  role         = "OWNER"
}

resource "masthead_data_domain" "example_domain" {
  name               = "Finance Domain"
  email              = "finance@example.com"
  slack_channel_name = "#finance-team"
}
