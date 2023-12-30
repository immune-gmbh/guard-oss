module V2
  UnauthorizedError = Class.new(StandardError)

  class ApiBaseController < ::ActionController::API
    include Traceable

    before_action :authenticate

    def current_ability
      @current_ability ||= Ability.new current_actor
    end

    def current_actor
      @current_actor ||= actor
    end

    def flush_actor
      @current_actor = nil
    end

    rescue_from V2::UnauthorizedError do
      render status: 401, json: { error: I18n.t('errors.unauthorized') }
    end

    rescue_from ActiveRecord::RecordNotFound do
      render status: 404, json: { error: I18n.t('errors.not_found') }
    end

    rescue_from CanCan::AccessDenied do
      render status: 403, json: { error: I18n.t('errors.unauthorized') }
    end

    # Expects a an hash mapping of an include-string to a condition. Similar to classnames utility.
    def conditional_include(include_to_condition)
      include_to_condition.map do |k, v|
        v ? k : nil
      end.flatten.compact
    end

    def feature?(feat)
      (Settings.features[feat] || []).include?(current_actor.organisation&.public_id)
    end

    def serialize_errors(model)
      @errors ||= []
      @errors += model.errors.to_hash.map do |attrib, msg|
        { id: attrib.to_s, title: msg.join(',') }
      end
    end

    private

    def authenticate
      raise V2::UnauthorizedError if current_actor.anonymous?
    end

    def actor
      span = current_span

      if request.headers.fetch('Authorization', '').start_with? 'Bearer '
        token = request.headers['Authorization'][7..]
        span.add_event "Trying auth token '#{token}'"

        begin
          act = TokenService.verify_token(token)
          if act.user?
            span.add_event "User auth with user ID '#{act.user.id}'"
          elsif act.organisation
            span.add_event "Service auth for organisation ID '#{act.organisation.id}'"
          else
            span.add_event "Service auth for '#{act.service}'"
          end
          return act
        rescue ActiveRecord::RecordNotFound => e
          span.add_event 'Membership unknown', attributes: { error: e }
        rescue TokenError => e
          span.add_event 'Token validation failed', attributes: { error: e }
        end
      end
      span.add_event 'Falling back to anonymous user'

      Actor.anonymous
    end
  end
end
