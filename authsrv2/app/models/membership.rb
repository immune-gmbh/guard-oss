class Membership < ActiveRecord::Base
  self.implicit_order_column = :created_at

  belongs_to :user
  belongs_to :organisation

  has_secure_token :jwt_token_key

  before_validation :set_jwt_token, if: -> { jwt_token_key.blank? }

  enum status: [:pending, :active, :suspended, :deleted]
  enum role: [:owner, :admin, :user], _prefix: true

  def set_jwt_token
    self.jwt_token_key = SecureRandom.hex if pending? || active?
  end
end
