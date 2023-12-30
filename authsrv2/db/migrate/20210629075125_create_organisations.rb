class CreateOrganisations < ActiveRecord::Migration[6.1]
  def change
    create_table :organisations, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.string :name
      t.string :vat_number
      t.references :subscription, type: :uuid
      
      t.timestamps
    end
  end
end
