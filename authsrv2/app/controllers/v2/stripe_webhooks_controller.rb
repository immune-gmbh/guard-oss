module V2
  class StripeWebhooksController < V2::ApiBaseController
    skip_before_action :authenticate

    def event
      begin
        payload = request.body.read
        sig_header = request.env['HTTP_STRIPE_SIGNATURE']
        event = nil

        event = Stripe::Webhook.construct_event(payload, sig_header, STRIPE_WEBHOOK_SIGNING_SECRET)
        Stripe::EventHandler.new.dispatch(event)
      rescue JSON::ParserError => e
        # Invalid payload
        return render status: 400
      rescue Stripe::SignatureVerificationError => e
        # Invalid signature
        return render status: 400, json: { error: e }
      end

      head :ok
    end
  end
end
