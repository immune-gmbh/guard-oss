class User < ActiveRecord::Base
  self.implicit_order_column = :created_at

  authenticates_with_sorcery!

  has_many :authentications, :dependent => :destroy
  accepts_nested_attributes_for :authentications

  has_many :memberships, dependent: :destroy
  has_many :organisations, through: :memberships, source: :organisation

  belongs_to :address, optional: true
  accepts_nested_attributes_for :address

  before_save { email.downcase! if email_changed? }
  before_validation :generate_random_password, if: -> { new_record? }

  validates :name, :email, presence: true
  validates :email, uniqueness: true
  validates_format_of :email, with: URI::MailTo::EMAIL_REGEXP

  before_update :setup_activation, if: -> { email_changed? }
  after_update :send_activation_needed_email!, if: -> { previous_changes["email"].present? }

  enum role: [:admin, :user]
  enum login_status: [:logged_in, :logged_out]

  def generate_random_password
    self.password = SecureRandom.hex
  end

  def regenerate_membership_token_keys
    # @TODO: this should be a batch update
    memberships.each do |membership|
      next unless membership.active?

      membership.set_jwt_token
      membership.save
    end
  end

  def self.find_by_id(id)
    self.find(id)
  end

  def prepare_token_login
    self.update(login_token: SecureRandom.hex)
  end

  def self.login_with_token(token)
    user = self.find_by(login_token: token)
    user&.update(login_token: nil)
    user
  end
end
