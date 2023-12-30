module V2
  class SubscriptionsController < V2::ApiBaseController
    load_and_authorize_resource :subscription, except: [:intent, :default_payment_method, :setup, :create]

    def intent
      organisation = Organisation.find(params[:organisation_id])
      authorize! :manage, organisation

      if Settings.payment.disable
        render status: :service_unavailable, json: { errors: [{ id: "settings", title: "Payments disabled" }] }
      else
        response = OrganisationService.setup_payment(organisation)

        if response.success
          render json: SubscriptionSerializer.new(response.subscription, {
            params: { setup_intent: response.setup_intent } }).serializable_hash
        else
          render status: :service_unavailable, json: { errors: [{ id: "stripe", title: response.message }] }
        end
      end
    end

    def default_payment_method
      organisation = Organisation.find(params[:organisation_id])
      authorize! :manage, organisation

      if Settings.payment.disable
        render status: :service_unavailable, json: { errors: [{ id: "settings", title: "Payments disabled" }] }
      else
        response =
          OrganisationService.make_payment_method_default(organisation,
                                                          params[:payment_method_id])

        if response.success
          render json: SubscriptionSerializer.new(response.subscription).serializable_hash
        else
          render status: :service_unavailable, json: { errors: [{ id: "stripe", title: response.message }] }
        end
      end
    end

    def create
      membership = Membership.find(membership_params[:membership_id])
      authorize! :manage, membership.organisation
      authorize! :read, membership.user

      response = OrganisationService.create_subscription(membership)
      if response.success
        render json: SubscriptionSerializer.new(response.subscription).serializable_hash
      else
        render status: :service_unavailable, json: { errors: [{ id: "stripe", title: response.message }] }
      end
    end

    def index
      render json: SubscriptionSerializer.new(@subscriptions).serializable_hash
    end

    def show
      render json: SubscriptionSerializer.new(@subscription).serializable_hash
    end

    private

    def membership_params
      params.permit(:membership_id)
    end
  end
end
