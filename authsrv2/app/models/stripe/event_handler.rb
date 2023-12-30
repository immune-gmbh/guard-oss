class Stripe::EventHandler
  def dispatch(event)
    method_name = "handle_#{event.type.gsub('.', '_')}"
    send(method_name, event) if respond_to?(method_name)
  end

  # Event: customer.subscription.created
  def handle_customer_subscription_created(event)
    stripe_subscription = event.data['object']

    subscription = Subscription.find_by!(stripe_subscription_id: stripe_subscription['id']) # check if its valid
    # inform orga main user of subscription start
    subscription.update!(status: 'active')
    SubscriptionMailer.created_email(subscription).deliver_now
  end

  # Event: customer.subscription.deleted
  def handle_customer_subscription_deleted(event)
    stripe_subscription = event.data['object']

    subscription = Subscription.find_by!(stripe_subscription_id: stripe_subscription['id'])
    OrganisationService.delete(organisation: subscription.organisation, soft_delete: true)
  end

   # Event: customer.subscription.updated, triggered on many occasions, like status or period changes
   # we use it primarily to update the current billing period of the subscription in the DB
  def handle_customer_subscription_updated(event)
    stripe_subscription = event.data['object']

    StripeService.update_subscription(stripe_subscription)
  end

  # Event: invoice.created, occurs 1 hour before finalization and billing attempt for subscriptions
  # A new billing period has started just now
  def handle_invoice_created(event)
    ActiveRecord::Base.transaction do
      stripe_invoice = event.data['object']
      StripeService.create_or_update_invoice(stripe_invoice)

      return if stripe_invoice['billing_reason'] == 'subscription_create'

      subscription = Subscription.find_by!(stripe_subscription_id: stripe_invoice['subscription'])
      # This ensures that at least one usage_record gets created for the new period in the DB and Stripe
      # see usage_records.rake
      subscription.update(new_devices_amount: subscription.current_devices_amount) unless subscription.new_devices_amount
    end
  end

  # Event: invoice.payment_succeeded
  def handle_invoice_paid(event)
    stripe_invoice = event.data['object']

    invoice = StripeService.create_or_update_invoice(stripe_invoice)

    return if stripe_invoice['billing_reason'] == 'subscription_create'

    InvoiceMailer.invoice_paid_email(invoice).deliver_now
  end

  # Event: invoice.payment_failed
  def handle_invoice_payment_failed(event)
    stripe_invoice = event.data['object']

    invoice = StripeService.create_or_update_invoice(stripe_invoice)

    InvoiceMailer.payment_failed_email(invoice).deliver_now
  end

  # Event: invoice.marked_uncollectible
  def handle_invoice_marked_uncollectible(event)
    stripe_invoice = event.data['object']

    invoice = StripeService.create_or_update_invoice(stripe_invoice)

    OrganisationService.delete(organisation: invoice.subscription.organisation, soft_delete: true)

    SubscriptionMailer.deactivate_email(invoice.subscription).deliver_now
  end

  # Event: invoice.finalized, sent just before automatic collection is being triggered
  def handle_invoice_finalized(event)
    stripe_invoice = event.data['object']

    StripeService.create_or_update_invoice(stripe_invoice)
  end

  # Event: invoice.voided
  def handle_invoice_voided(event)
    stripe_invoice = event.data['object']

    # deactivate subscription inform orga-mailer
  end
end
