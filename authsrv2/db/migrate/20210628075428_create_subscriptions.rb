class CreateSubscriptions < ActiveRecord::Migration[6.1]
  def change
    create_table :subscriptions, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.references :user, type: :uuid
      t.string :stripe_subscription_id, null: true
      t.string :stripe_subscription_item_id, null: true
      t.string :status, default: 'created', null: false
      t.integer :current_devices_amount, null: false, default: 0
      t.integer :new_devices_amount
      t.boolean :notify_invoices, default: false, null: false
      t.boolean :notify_device_updates, default: false, null: false
      t.float :tax_rate
      t.date :period_start
      t.date :period_end

      t.timestamps
    end
  end
end
