class StripeService < ApplicationService

  def create_customer
    # TODO: Create Stripe Customer here
  end

  def self.create_or_update_invoice(stripe_invoice)
    # update our invoice 
    Invoice.transaction do
      invoice = Invoice.find_or_create_by(stripe_invoice_id: stripe_invoice['id'])

      subscription = Subscription.find_by(stripe_subscription_id: stripe_invoice['subscription'])
      invoice.subscription = subscription
      invoice.stripe_invoice_number ||= stripe_invoice['number']
      invoice.finalized_at ||= Time.at(stripe_invoice['status_transitions']['finalized_at']) if stripe_invoice['status_transitions']['finalized_at']
      invoice.paid_at ||= Time.at(stripe_invoice['status_transitions']['paid_at']) if stripe_invoice['status_transitions']['paid_at']
      invoice.marked_uncollectible_at ||= Time.at(stripe_invoice['status_transitions']['marked_uncollectible_at']) if stripe_invoice['status_transitions']['marked_uncollectible_at']
      invoice.voided_at ||= Time.at(stripe_invoice['status_transitions']['voided_at']) if stripe_invoice['status_transitions']['voided_at']
      invoice.tax_rate ||= stripe_invoice['tax_rate']
      invoice.subtotal ||= stripe_invoice['subtotal']
      invoice.total ||= stripe_invoice['total']
      # don't downgrade invoice status from paid, uncollectible or void to open or draft:
      invoice.status = stripe_invoice['status'] unless ['paid', 'marked_uncollectible', 'voided'].include?(invoice.status)
      invoice.stripe_pdf_url ||= stripe_invoice['invoice_pdf']
      invoice.save!

      invoice
    end
  end

  def self.update_subscription(stripe_subscription)
    Subscription.transaction do
      subscription = Subscription.find_by!(stripe_subscription_id: stripe_subscription['id'])

      subscription.period_start = Time.at(stripe_subscription['current_period_start']).to_date
      subscription.period_end = Time.at(stripe_subscription['current_period_end']).to_date
      subscription.status = stripe_subscription['status']

      subscription.save!
    end
  end

  def self.update_usage(subscription)
    # This method calculates amount to be billed at the end of the period if no further changes to the
    # enrolled devices are made. If changes occur, this method will be called again to reflect that.
    # The underlying Stripe Product/Price will only use the last submitted record in each period for the
    # creation of the invoice and the amount to be billed.
    # It takes the first usage_record (reported # of devices in use at a certain date) of
    # the active period and multiplies it by the number of days up to the next usage_record or
    # period_end, after which it is divided by the number of days in current period.
    # The result is a fair pricing model which accounts for daily variations of enrolled devices
    # e.g.:
    # (amount1 * days1 + amount2 * days2...amountN * daysN) / daysTotal
    # Example1: (10 Devices * 30 days) / 30 days = 10 Devices to be billed for the month
    # Example2: (5 Devices * 20 days + 15 devices * 10 days) / 30 days
    #   = 250 DeviceDays / 30 days => ~8 Devices to be billed for the month

    period_start = subscription.period_start
    period_end = subscription.period_end

    Subscription.transaction do
      subscription = Subscription
        .where(stripe_subscription_id: subscription.stripe_subscription_id)
        .includes(:usage_records)
        .where('usage_records.date >= ? AND usage_records.date < ?', period_start, period_end)
        .references(:usage_records)
        .first

      total_sum = 0
      usage_records = subscription.usage_records.sort_by(&:date)
      usage_records.each_with_index do |_record, index|
        days = ((index == usage_records.length - 1 ? period_end : usage_records[index + 1].date) - usage_records[index].date).to_i
        total_sum += (usage_records[index].amount * days)
      end

      # Submit full device months to the specific stripe subscription item.
      begin
        Stripe::SubscriptionItem.create_usage_record(
          subscription.stripe_subscription_item_id,
          {
            quantity: subscription.organisation.freeloader ? 0 : (total_sum / (period_end - period_start)).round,
            action: 'set',
            timestamp: DateTime.now.to_i
          }
        )
        # create a new usage_record for start of new billing period unless already present
        subscription.usage_records.find_or_create_by(date: period_end) do |usage_record|
          usage_record.amount = subscription.current_devices_amount
        end
      rescue Stripe::StripeError => e
        SubscriptionMailer.usage_record_update_error_email(subscription, e).deliver_now
      end
    end
  end
end
