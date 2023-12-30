module V2
  class UsersController < V2::ApiBaseController
    load_and_authorize_resource :user, except: [:activate, :create, :resend]
    skip_before_action :authenticate, only: [:create, :activate, :resend]

    def index
      render json: UserSerializer.new(@users, { include: [params.permit(:include)[:include]] }).serializable_hash.to_json
    end

    def create
      if !user_params[:password] || user_params[:password] == ""
        render status: :unprocessable_entity, json: { errors: [ id: "password", title: "empty password" ] }
        return
      end
      
      ActiveRecord::Base.transaction do
        user = User.create(user_params)

        if user.persisted?
          if Settings.authentication.disable_activation then
            user.activate!

            response = OrganisationService.create_organisation_and_subscription(user, {})
            if !response.success
              render status: :service_unavailable, json: { errors: [ { id: "activate", title: response.message } ] }
              raise ActiveRecord::Rollback
            end
          else
            UserMailer.activation_needed_email(user).deliver_now
          end

          render json: UserSerializer.new(user).serializable_hash.to_json
        else
          render status: :unprocessable_entity, json: { errors: serialize_errors(user) }
        end
      end
    end

    def activate
      user = User.load_from_activation_token(params[:id])
      if user
        ActiveRecord::Base.transaction do
          user.activate!

          response = OrganisationService.create_organisation_and_subscription(user, {})

          if response.success
            user.reload
            current_ability = Ability.new(Actor.new(:user, user: user))

            session = Session.new
            session.user = user
            session.memberships = user.memberships
            session.default_organisation = user.organisations.first&.public_id

            render json: SessionSerializer.new(session, { params: { current_ability: current_ability } })
          else
            render status: :service_unavailable, json: { errors: [ { id: "activate", title: response.message } ] }
            raise ActiveRecord::Rollback
          end
        end
      else
        render status: :not_found, json: { errors: [ { id: "id", title: "Activation token wasn't found" } ] }
      end
    end

    def resend
      # XXX rate limit
      user = User.find_by(email: params[:email].downcase)
      if user && user.activation_state == "pending"
        UserMailer.activation_needed_email(user).deliver_now
      end
      # Do not allow information exfiltration, same response whether the email exist or not
      render json: { success: true }
    end

    def show
      render json: UserSerializer.new(@user, include: [:address]).serializable_hash.to_json
    end

    def update
      # update user
      if !@user.update(user_params.merge(user_admin_params))
        render status: :unprocessable_entity, json: { errors: serialize_errors(@user) }
        return
      end

      # XXX update address
      render json: UserSerializer.new(@user, include: [:address]).serializable_hash.to_json
    end

    def change_password
      ActiveRecord::Base.transaction do
        user = if current_actor.user.crypted_password.present?
          User.authenticate(current_actor.user.email, user_params[:current_password])
        else
          current_actor.user
        end

        return render json: { success: false, message: "User not found or wrong password." } unless user.present?

        success = user.change_password!(user_params[:password])

        return render json: { success: success }
      end
    end

    def destroy
      EventService.new.revoke_token(@user.memberships)
      @user.destroy

      render json: { success: !@user.persisted? }
    end

    private

    def address_params
      params.permit(address: [:street_and_number, :postal_code, :city, :country])
    end

    def user_params
      params.permit(:name, :email, :current_password, :password, :password_confirmation, :has_seen_intro)
    end

    def user_admin_params
      return {} unless current_actor.user? && current_actor.user.admin?

      params.permit(:role)
    end
  end
end
