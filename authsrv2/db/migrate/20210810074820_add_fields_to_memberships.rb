class AddFieldsToMemberships < ActiveRecord::Migration[6.1]
  def change
    add_column :memberships, :notify_device_update, :boolean, default: false
    add_column :memberships, :notify_invoice, :boolean, default: false
  end
end
