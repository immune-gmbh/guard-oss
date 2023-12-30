module V2
  class OauthsController < V2::ApiBaseController
    skip_before_action :authenticate

    # sends the user on a trip to the provider,
    # and after authorizing there back to the callback url.
    def oauth
      login_at(params[:provider])
    end

    def callback
      provider = params[:provider]
      base = Settings.external_frontend_url
      if @user = login_from(provider)
        @user.prepare_token_login

        redirect_to "#{base}/login?token=#{@user.login_token}"
      else
        ActiveRecord::Base.transaction do
          if @user = User.find_by(email: @user_hash[:user_info]['email'].downcase)
            @user.add_provider_to_user(provider, @user_hash[:uid])
            @user.prepare_token_login

            redirect_to "#{base}/login?token=#{@user.login_token}"
          elsif Settings.authentication.disable_registration
            render status: 403, json: { errors: [{ id: "registration disabled" }] }
          else
            @user = create_from(provider)
            # NOTE: this is the place to add '@user.activate!' if you are using user_activation submodule
            @user.setup_activation
            @user.save

            redirect_to "#{base}/registration/confirm?activationToken=#{@user.activation_token}"
          end
        end
      end
    end
  end
end
