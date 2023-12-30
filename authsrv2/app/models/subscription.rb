class Subscription < ActiveRecord::Base
  self.implicit_order_column = :created_at

  belongs_to :organisation

  has_many :invoices
  has_many :usage_records

  # TODO Fill mock
  def max_devices_amount
    Settings.payment.device_quota
  end

  # Base Fee in Dollar
  def monthly_base_fee
    15
  end

  def monthly_fee_per_device
    5
  end
end
