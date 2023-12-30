class SubscriptionSerializer < BaseSerializer
  attributes :status, :current_devices_amount, :max_devices_amount, :period_start, :period_end, :tax_rate, :monthly_base_fee, :monthly_fee_per_device

  attribute :action_required do |object|
    # check if the last finalized invoice is already paid 1 hour after finalization
    object.status == "past_due"
  end

  attribute :billing_details do |object|
    if Settings.payment.disable
      {}
    else
      stripe_customer = Stripe::Customer.retrieve({id: object.organisation.stripe_customer_id, expand: ["invoice_settings.default_payment_method"]})

      default_payment_method = stripe_customer.try(:invoice_settings).try(:default_payment_method)

      if default_payment_method then
        {
          last4: default_payment_method.present? ? "****#{default_payment_method.card.last4}" : nil,
          expiry_date: default_payment_method.present? ? "#{default_payment_method.card.exp_month}/#{default_payment_method.card.exp_year}" : nil,
          free_credits: -stripe_customer.balance,
        }
      else
        { free_credits: 0 }
      end
    end
  end

  attribute :setup_intent_secret do |object, params|
    params[:setup_intent]&.client_secret
  end
end
