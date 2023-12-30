class AddAdditionalFieldsToOrganisations < ActiveRecord::Migration[6.1]
  def change
    add_column :organisations, :stripe_customer_id, :string
    add_column :organisations, :contact, :string
    add_column :organisations, :invoice_email_address, :string
    add_column :organisations, :invoice_name, :string
    add_column :organisations, :splunk_enabled, :boolean
    add_column :organisations, :splunk_event_collector_url, :string
    add_column :organisations, :splunk_authentication_token, :string
    add_column :organisations, :syslog_enabled, :boolean
    add_column :organisations, :syslog_hostname_or_address, :string
    add_column :organisations, :syslog_udp_port, :string
  end
end
