class Invoice < ActiveRecord::Base
  self.implicit_order_column = :created_at

  belongs_to :subscription

  validates_uniqueness_of :stripe_invoice_id
end
