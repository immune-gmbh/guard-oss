class MembershipSerializer < BaseSerializer
  belongs_to :organisation
  belongs_to :user

  attributes :role, :notify_device_update, :notify_invoice

  abilities :delete

  attribute :token do |membership|
    TokenService.issue_api_token(membership)
  end

  attribute :enrollment_token, object_method_name: :active? do |membership|
    next unless membership.active? || membership.organisation.active?

    TokenService.issue_enrollment_token(membership)
  end

  def self.attribute_names 
    [:organisation,:user,:role,:can_delete,:token,:enrollment_token] + BaseSerializer.attribute_names
  end
end
