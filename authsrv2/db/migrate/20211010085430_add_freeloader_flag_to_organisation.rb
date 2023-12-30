class AddFreeloaderFlagToOrganisation < ActiveRecord::Migration[6.1]
  def change
    add_column :organisations, :freeloader, :boolean, default: false, null: false
  end
end
