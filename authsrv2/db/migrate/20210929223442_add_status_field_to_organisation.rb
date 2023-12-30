class AddStatusFieldToOrganisation < ActiveRecord::Migration[6.1]
  def change
    add_column :organisations, :status, :integer, default: 0 # created
  end
end
