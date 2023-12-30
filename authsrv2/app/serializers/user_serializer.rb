class UserSerializer < BaseSerializer
  attributes :name, :email, :invited, :role, :activation_state, :has_seen_intro

  has_many :organisations
  belongs_to :address

  attribute :has_password do |user|
    user.crypted_password.present?
  end

  attribute :authenticated_google do |user|
    user.authentications.map(&:provider).include? "google"
  end

  attribute :authenticated_github do |user|
    user.authentications.map(&:provider).include? "github"
  end

  def self.attribute_names
    [:name, :email, :invited, :role, :activation_state, :has_seen_intro, :organisations, :address, :has_password, :authenticated_github, :authenticated_google] + BaseSerializer.attribute_names
  end
end
