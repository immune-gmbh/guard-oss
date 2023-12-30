class SubscriptionMailer < ApplicationMailer
  def created_email(subscription)
    @user = subscription.organisation.memberships.find_by(status: :active, role: :owner).user
    @organisation = subscription.organisation

    mail(to: @user.email, subject: "New Subscription for #{@organisation.name}")
  end

  def deactivated_email(subscription)
    @organisation = invoice.subscription.organisation
    subscription.organisation.memberships.role_owner.each do |membership|
      @user =  membership.user
      mail(to: @user.email, subject: "Deactivated Subscription for #{@organisation.name}")
    end
  end

  def usage_record_update_error_email(subscription, error)
    @subscription = subscription
    @error = error
    mail(to: "team@immune.gmbh", subject: "A usage record could not be transmitted to Stripe")
  end
end
