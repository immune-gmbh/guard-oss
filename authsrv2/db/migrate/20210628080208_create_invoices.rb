class CreateInvoices < ActiveRecord::Migration[6.1]
  def change
    create_table :invoices, id: false do |t|
      t.uuid :public_id, primary_key: true, default: -> { "next_uuid_seq('uuid_comb_sequence'::text)" }
      t.references :subscription, type: :uuid
      t.string :stripe_invoice_id
      t.string :stripe_invoice_number
      t.string :stripe_pdf_url
      t.datetime :finalized_at
      t.datetime :paid_at
      t.datetime :marked_uncollectible_at
      t.datetime :voided_at
      t.float :tax_rate
      t.integer :subtotal
      t.integer :total
      t.string :status

      t.timestamps
    end
  end
end
