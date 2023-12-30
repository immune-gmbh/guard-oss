class Organisation < ActiveRecord::Base
  self.implicit_order_column = :created_at

  has_one :subscription, dependent: :destroy
  belongs_to :address, optional: true, dependent: :destroy
  accepts_nested_attributes_for :address, update_only: true

  has_many :memberships, dependent: :destroy
  has_many :users, through: :memberships

  validates :name, presence: true, uniqueness: true
  validates :vat_number, valvat: {
    match_country: :country_code,
    allow_blank: true,
    lookup: if Rails.env.production? then { raise_error: false } else false end,
  }, if: :vat_number_changed? && -> { address&.public_id && country_code != "GB" }

  # "Soft Delete", .delete!, .delete
  enum status: [:created, :active, :suspended, :deleted]

  before_save :mark_for_usage_record_creation, if: :freeloader_changed? && :persisted?

  def owner
    self.memberships.owner
  end

  def country_code
    ISO3166::Country.find_country_by_name(address&.country)&.alpha2
  end

  def mark_for_usage_record_creation
    if self.subscription.present?
      self.subscription.update(new_devices_amount: self.subscription.current_devices_amount) unless self.subscription.new_devices_amount
    end
  end
end
