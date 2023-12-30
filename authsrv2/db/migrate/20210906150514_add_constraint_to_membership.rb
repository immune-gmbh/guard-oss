class AddConstraintToMembership < ActiveRecord::Migration[6.1]
  def change
    change_column_null :memberships, :notify_invoice, true
    change_column_null :memberships, :notify_device_update, true
  end
end
