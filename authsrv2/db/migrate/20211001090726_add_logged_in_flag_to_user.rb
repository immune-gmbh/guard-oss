class AddLoggedInFlagToUser < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :login_status, :integer, default: 0, null: false
  end
end
