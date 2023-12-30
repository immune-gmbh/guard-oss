class CreateUsageRecords < ActiveRecord::Migration[6.1]
  def change
    create_table :usage_records do |t|
      t.references :subscription, type: :uuid, null: false
      t.integer :amount, null: false
      t.date  :date, null: false

      t.timestamps
    end
  end
end
