variables {
  api_token = "dummy-token-for-testing"
}

// Test the creation of all resources
run "create_all_resources" {
  command = plan

  assert {
    condition     = masthead_user.example_user.email == "user@example.com"
    error_message = "User email doesn't match expected value"
  }

  assert {
    condition     = masthead_user.example_user.role == "USER"
    error_message = "User role doesn't match expected value"
  }

  assert {
    condition     = masthead_data_domain.example_domain.name == "Test Domain"
    error_message = "Data domain name doesn't match expected value"
  }

  assert {
    condition     = masthead_data_domain.example_domain.email == "test@example.com"
    error_message = "Data domain email doesn't match expected value"
  }

  assert {
    condition     = masthead_data_domain.example_domain.slack_channel_name == "#10x-infra"
    error_message = "Data domain slack channel doesn't match expected value"
  }

  assert {
    condition     = masthead_data_product.example.name == "Test Data Product"
    error_message = "Data product name doesn't match expected value"
  }

  assert {
    condition     = masthead_data_product.example.description == "Product containing company analytics data"
    error_message = "Data product description doesn't match expected value"
  }

  assert {
    condition     = length(masthead_data_product.example.data_assets) == 1
    error_message = "Expected exactly one data asset"
  }

  assert {
    condition     = masthead_data_product.example.data_assets[0].type == "TABLE"
    error_message = "Data asset type doesn't match expected value"
  }

  assert {
    condition     = masthead_data_product.example.data_assets[0].project == "httparchive"
    error_message = "Data asset project doesn't match expected value"
  }
}

// Test the apply lifecycle
run "apply_test" {
  command = apply

  // This will use the plan from the first run
  variables {
    api_token = "dummy-token-for-testing"
  }

  assert {
    condition     = masthead_data_domain.example_domain.uuid != ""
    error_message = "Data domain UUID should not be empty after apply"
  }

  assert {
    condition     = masthead_data_product.example.data_domain_uuid == masthead_data_domain.example_domain.uuid
    error_message = "Data product's domain UUID doesn't match the domain's UUID"
  }
}

// Test for resource modification
run "modify_resources" {
  command = plan

  variables {
    api_token = "dummy-token-for-testing"
  }

  // Modify resources for testing updates
  module {
    source = "."
  }

  override_resource {
    target = masthead_data_domain.example_domain
    values = {
      name = "Updated Domain Name"
    }
  }

  override_resource {
    target = masthead_data_product.example
    values = {
      description = "Updated product description"
    }
  }

  assert {
    condition     = masthead_data_domain.example_domain.name == "Updated Domain Name"
    error_message = "Data domain name wasn't updated correctly"
  }

  assert {
    condition     = masthead_data_product.example.description == "Updated product description"
    error_message = "Data product description wasn't updated correctly"
  }
}
