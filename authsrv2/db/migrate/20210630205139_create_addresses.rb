class CreateAddresses < ActiveRecord::Migration[6.1]
  def change
    create_table :addresses, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.string :street_and_number
      t.string :city
      t.string :postal_code
      t.string :country

      t.timestamps
    end
  end
end
