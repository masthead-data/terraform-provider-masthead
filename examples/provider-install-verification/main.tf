terraform {
  required_providers {
    hashicups = {
      source = "hashicorp.com/edu/masthead-data"
    }
  }
}

provider "hashicups" {}

data "hashicups_coffees" "example" {}
