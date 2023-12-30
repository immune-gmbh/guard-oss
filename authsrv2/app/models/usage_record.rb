class UsageRecord < ActiveRecord::Base
  self.implicit_order_column = :created_at
  
  belongs_to :subscription

  validates_presence_of :amount
  validates_uniqueness_of :date, scope: :subscription
end
