if !Rails.env.test?
  Stripe.api_key = ENV.fetch "AUTHSRV_STRIPE_SECRET_KEY" do
    Rails.application.credentials.stripe[:secret_key]
  end

  STRIPE_WEBHOOK_SIGNING_SECRET = ENV.fetch "AUTHSRV_STRIPE_WEBHOOK_SIGNING_SECRET" do
    Rails.application.credentials.dig(:stripe, :webhook_secret)
  end
end
