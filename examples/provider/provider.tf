terraform {
  required_providers {
    masthead = {
      source = "masthead-data/masthead"
    }
  }
}

variable "api_token" {
  type      = string
  sensitive = true
}

provider "masthead" {
  api_token = var.api_token
}

resource "masthead_user" "example_user" {
  email = "user@example.com"
  role  = "USER"
}

resource "masthead_data_domain" "example_domain" {
  name  = "Test Domain"
  email = "test@example.com"
  slack_channel = {
    name = "10x-infra"
  }
}

resource "masthead_data_product" "example" {
  name        = "Test Data Product"
  description = "Product containing company analytics data"
  domain = {
    uuid = masthead_data_domain.example_domain.uuid
  }

  data_assets = [{
    type    = "TABLE"
    project = "httparchive"
    dataset = "crawl"
    table   = "pages"
  }]
}
