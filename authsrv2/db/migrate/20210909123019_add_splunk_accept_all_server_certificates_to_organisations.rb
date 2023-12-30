class AddSplunkAcceptAllServerCertificatesToOrganisations < ActiveRecord::Migration[6.1]
  def change
    add_column :organisations, :splunk_accept_all_server_certificates, :boolean, default: true
  end
end
