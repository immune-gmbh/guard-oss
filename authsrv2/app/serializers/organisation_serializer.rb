class OrganisationSerializer < BaseSerializer
  attributes :name, :vat_number, :splunk_enabled, :splunk_accept_all_server_certificates, :splunk_event_collector_url, :splunk_authentication_token, :syslog_enabled, :syslog_hostname_or_address, :syslog_udp_port, :invoice_name, :status, :freeloader

  attribute :member_count do |object|
    object&.memberships&.size || 0
  end

  has_many :memberships

  has_many :users do |object|
    object.users
  end

  has_one :subscription

  belongs_to :address

  abilities :edit, :delete, :read

  def self.attribute_names
    [ :name, :vat_number, :splunk_enabled, :splunk_accept_all_server_certificates, :splunk_event_collector_url, :splunk_authentication_token, :syslog_enabled, :syslog_hostname_or_address, :syslog_udp_port, :invoice_name, :status, :freeloader, :memberships, :users, :subscription, :address, :can_edit, :can_delete, :can_read, :member_count ] + BaseSerializer.attribute_names
  end
end
