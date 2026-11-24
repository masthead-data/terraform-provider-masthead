# The scan table (project.dataset.table) holds table_reference, rule_name, timestamp and value rows,
# usually refreshed by a scheduled query. Masthead checks each closed day 6 hours after it ends.
resource "masthead_data_scan" "orders_nulls" {
  name                = "Orders: nulls and duplicates"
  description         = "Daily null and duplicate rates of the orders tables, per country"
  table               = "my-project.masthead_dq.orders_columns"
  frequency           = "DAILY"
  process_delay_hours = 6
}

# Fixed limits instead of a learned range: each row of the scan table also has lower_threshold and
# upper_threshold (an empty one means no limit on that side). Incidents get severity P2.
resource "masthead_data_scan" "orders_null_limits" {
  name        = "Orders: null rate limits"
  description = "Null rate at most 3% per column; a breach is a P2 incident"
  table       = "my-project.masthead_dq.orders_null_limits"
  frequency   = "DAILY"
  type        = "ANALYSIS_WITH_THRESHOLD"
  severity    = "P2"
  labels = {
    managed_by = "terraform"
    owner      = "data-ops"
  }
}

# A schedule instead of a fixed frequency: the cron's ticks cut the periods. This source isn't loaded on
# Saturdays, so Friday's period runs from Friday 00:00 to Sunday 00:00 (Paris time) and no Saturday is missing.
resource "masthead_data_scan" "store_sales" {
  name              = "Store sales: daily volume"
  table             = "my-project.masthead_dq.store_sales_volume"
  schedule          = "0 0 * * SUN-FRI"
  schedule_timezone = "Europe/Paris"
}
