class SessionSerializer < BaseSerializer
  attributes :next_path, :default_organisation

  has_one :user
  has_many :memberships

  def initialize(session, options)
    options ||= {}
    options[:include] ||= [
      :user,
      :memberships,
      "user.organisations",
      "user.organisations.address",
      "user.organisations.subscription",
      "user.organisations.subscription.billingDetails"
    ]
    # don't include user relationship in membership and organisation to
    # prevent cyclic object references.
    options[:fields] ||=  {
      membership: MembershipSerializer.attribute_names - [:user],
      organisation: OrganisationSerializer.attribute_names - [:users, :memberships],
      user: UserSerializer.attribute_names
    }
    super(session, options)
  end

end
