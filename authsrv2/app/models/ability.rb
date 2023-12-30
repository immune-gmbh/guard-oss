# frozen_string_literal: true

class Ability
  include CanCan::Ability

  def initialize(actor)
    if actor.user?
      if actor.user.admin?
        can :manage,              :all
      else
        organisation_owner = {
          organisation_id: actor.owner_of.map(&:organisation_id)
        }
        organisation_member = {
          organisation_id: actor.member_of.map(&:organisation_id)
        }

        can :manage,              User,         public_id: actor.user.id
        can :read,                User,         public_id: actor.owner_of.map(&:organisation).flat_map(&:memberships).map(&:user_id)
        can :read,                Subscription, organisation_member
        can :manage,              Subscription, organisation_owner
        can :read,                UsageRecord,  subscription: organisation_owner
        can [:read, :download],   Invoice,      subscription: organisation_owner

        can [:read, :delete],     Membership,   { user_id: actor.user.id, status: :active, organisation: { status: [:active, :created] }} # every user can see and delete his own memberships
        can :manage,              Membership,   { organisation_id: actor.owner_of.map(&:organisation_id), status: [:pending, :active], organisation: { status: [:created, :active] }} #every owner can manage all pending/active memberships of his created/active orgs

        can :create,              Organisation
        can :read,                Organisation, { public_id: actor.member_of.map(&:organisation_id), status: [:created, :active] }
        can :delete,              Organisation, { public_id: actor.owner_of.map(&:organisation_id), status: [:created, :active] }
        can :manage,              Organisation, { public_id: actor.owner_of.map(&:organisation_id), status: [:created, :active] }
      end
    elsif actor.service?
      can :notify, :all

      if actor.organisation
        can [:alert, :bill], Organisation, public_id: actor.organisation.id
      else
        can [:alert, :bill], Organisation
      end
    end
  end
end
