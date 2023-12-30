class Address < ActiveRecord::Base
  self.implicit_order_column = :created_at

  has_many :users
  has_many :organisations
  has_many :subscriptions

  validates :street_and_number, :city, :postal_code, :country, presence: true
end
