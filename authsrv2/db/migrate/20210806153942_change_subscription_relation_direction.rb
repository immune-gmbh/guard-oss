class ChangeSubscriptionRelationDirection < ActiveRecord::Migration[6.1]
  def up
    add_reference :subscriptions, :organisation, type: :uuid
    Organisation.all.each do |org|
      org.subscription&.update(organisation: org)
    end
    remove_reference :organisations, :subscription
  end

  def down
    add_reference :organisations, :subscription, type: :uuid
    Subscription.all.each do |sub|
      sub.organisation&.update(subscription: sub)
    end
    remove_reference :subscriptions, :organisation
  end
end
