class SorceryExternal < ActiveRecord::Migration[6.1]
  def change
    create_table :authentications do |t|
      t.uuid :user_id, null: false
      t.string :provider, :uid, null: false

      t.timestamps              null: false
    end

    add_index :authentications, [:provider, :uid]
  end
end
