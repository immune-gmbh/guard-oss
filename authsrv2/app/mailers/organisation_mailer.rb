class OrganisationMailer < ApplicationMailer
  def device_usage_will_update_email(organisation)
    @organisation = organisation
    @organisation.memberships.role_owner.each do |membership|
      next unless membership.notify_device_update

      @user = membership.user

      mail(to: @user.email, subject: "The Device Usage will be updated for #{@organisation.name}")
    end
  end

  def devices_updated_email(organisation)
    @organisation = organisation

    subscription = @organisation.subscription
    @current_devices_amount = subscription.current_devices_amount
    @new_devices_amount = subscription.new_devices_amount

    stripe_subscription = Stripe::Subscription.retrieve(subscription.stripe_subscription_id)
    device_price = stripe_subscription.plan.amount_decimal
    estimated_cost = device_price.to_f * @new_devices_amount.to_f
    @price = estimated_cost > 0 ? Money.from_cents(estimated_cost, 'EUR').format : nil

    stripe_customer = Stripe::Customer.retrieve({id: @organisation.stripe_customer_id, expand: ["invoice_settings.default_payment_method"]})
    @free_credits = stripe_customer.balance < 0 ? Money.from_cents(stripe_customer.balance.abs, 'EUR').format : nil

    @organisation.memberships.role_owner.each do |membership|
      next unless membership.notify_device_update
      @user = membership.user

      mail(to: @user.email, subject: "The Device Usage has been updated for #{@organisation.name}")
    end
  end

  def payment_reminder_email(organisation)
    @organisation = organisation
    @organisation.memberships.role_owner.each do |membership|
      @user = membership.user

      mail(to: @user.email, subject: "Payment Reminder")
    end
  end

  def deleted_email(organisation)
    @organisation = organisation

    @organisation.memberships.role_owner.each do |membership|
      @user = membership.user

      mail(to: @user.email, subject: "#{@organisation.name} has been deleted")
    end
  end
end
