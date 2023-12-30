module V2
  class SessionsController < V2::ApiBaseController
    skip_before_action :authenticate, only: [:create, :destroy, :refresh]

    def show
      render_session
    end

    def create
      if params[:token]
        user = User.login_with_token(params[:token])
      else
        user = User.authenticate(params[:email].downcase, params[:password])
      end
      raise V2::UnauthorizedError unless user

      EventService.new.revoke_token(user.memberships)
      user.regenerate_membership_token_keys

      if user.activation_state == 'pending' && user.address
        render json: { errorMessage: "Please confirm your email address", status: 403 }, status: 403
        return
      end
      base = Settings.external_frontend_url
      path = if user.activation_state == 'pending' && !user.address
        "#{base}/registration/activate_email?email=#{user.email}"
      elsif !user.has_seen_intro
        "#{base}/dashboard/welcome"
      else
        "#{base}/dashboard"
      end

      user.logged_in!
      current_actor = Actor.new(:user, user: user)
      current_ability = Ability.new(current_actor)

      session = Session.new
      session.next_path = path
      session.user = user
      session.memberships = user.memberships
      session.default_organisation = user.organisations.first&.public_id

      render json: SessionSerializer.new(session, { params: { current_ability: current_ability } })
    end

    def destroy
      if current_actor && current_actor.user.present?
        current_actor.user.logged_out!
        EventService.new.revoke_token(current_actor.user.memberships)
        current_actor.user.regenerate_membership_token_keys
      end

      render json: { ok: true }
    end

    def refresh
      token = (request.headers['Authorization'] || "")[7..]
      claims = JWT.decode token, nil, false
      claims &&= claims[0]
      jti = claims.fetch('jti', '')
      membership = Membership.find_by!(jwt_token_key: jti)

      current_actor = Actor.new(:user, user: membership.user)
      current_ability = Ability.new(current_actor)

      if current_actor.user.logged_in?
        render_session
      else
        raise V2::UnauthorizedError
      end
    rescue JWT::DecodeError
      raise V2::UnauthorizedError
    end

    private

    def render_session
      session = Session.new
      session.user = current_actor.user
      session.memberships = current_actor.user&.memberships || []
      session.default_organisation = current_actor.user&.organisations&.first&.public_id

      render json: SessionSerializer.new(session, { params: { current_ability: current_ability } })
    end
  end
end
