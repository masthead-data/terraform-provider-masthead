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

resource "masthead_user" "example_user1" {
  email = "user1@example.com"
  role  = "OWNER"
}

resource "masthead_data_domain" "example_domain1" {
  name  = "Test Domain1"
  email = "test@example.com"
}

resource "masthead_data_domain" "example_domain2" {
  name               = "Test Domain2"
  email              = "test1@example.com"
  slack_channel_name = "data-ops"
}

resource "masthead_data_product" "example_product1" {
  name             = "Test Data Product1"
  description      = "Product containing company analytics data"
  data_domain_uuid = masthead_data_domain.example_domain1.uuid
  data_assets = [{
    type    = "TABLE"
    project = "project_id"
    dataset = "dataset_id"
    table   = "table_id"
  }]
}
