class CreateMemberships < ActiveRecord::Migration[6.1]
  def change
    create_table :memberships, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.references :user, type: :uuid
      t.references :organisation, type: :uuid
      t.integer :role, null: false, default: 2 # user
      t.integer :status, null: false, default: 0 # pending
      t.string :jwt_token_key

      t.timestamps
    end

    add_index :memberships, [:user_id, :organisation_id], unique: true
  end
end
