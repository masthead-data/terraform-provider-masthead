# The scan table (project.dataset.table) holds table_reference, rule_name, timestamp and value rows,
# usually refreshed by a scheduled query. Masthead checks each closed day 6 hours after it ends.
resource "masthead_data_scan" "orders_nulls" {
  name                = "Orders: nulls and duplicates"
  description         = "Daily null and duplicate rates of the orders tables, per country"
  table               = "my-project.masthead_dq.orders_columns"
  frequency           = "DAILY"
  process_delay_hours = 6
}
