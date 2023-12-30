class CreateUser < ActiveRecord::Migration[6.1]
  def change
    create_table :users, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.string :name
      t.string :email, null: false
      t.string :crypted_password
      t.string :salt
      t.string :jwt_token_key
      t.integer :role, default: 1 # user
      t.boolean :invited, default: false

      t.timestamps null: false
    end

    add_index :users, :email, unique: true
  end
end
