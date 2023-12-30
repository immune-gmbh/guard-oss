# Actors are the unifying abstraction of (human) users and services. Users
# authorize using password and cookies against the database. Services authorize
# with self-issued Bearer JWT's and act on behalf of a organisation (group of
# users). A special anonymous Actor represents users that are not logged in.
#
# Optionaly, service tokens can have scope that further limits what they
# actions they authorize.
#
# See the Token Service for the machinery that issues and verifies to
# tokens.
class Actor
  attr_reader :user, :organisation, :service, :membership

  def initialize(type, user: nil, membership: nil, service: nil, organisation: nil)
    unless [:service, :user, :anonymous].include? type
      throw ArgumentError.new("unknown Actor type #{type}")
    end

    if (type == :user) && !user && !membership
      throw ArgumentError.new('missing user record')
    end

    if (type == :user) && (organisation || service)
      throw ArgumentError.new('user actors are scoped by membership')
    end

    if (type == :service) && !service
      throw ArgumentError.new('missing service name')
    end

    @type = type
    @membership = membership
    @service = service
    @user = user || membership&.user
    @organisation = organisation || membership&.organisation
  end

  def self.anonymous
    Actor.new(type = :anonymous)
  end

  def anonymous?
    @type == :anonymous
  end

  def user?
    @type == :user
  end

  def service?
    @type == :service
  end

  # Memberships w/ role == :owner
  def owner_of
    if user?
      @user.memberships.role_owner
    else
      []
    end
  end

  # Memberships
  def member_of
    if user?
      @user.memberships
    else
      []
    end
  end
end
