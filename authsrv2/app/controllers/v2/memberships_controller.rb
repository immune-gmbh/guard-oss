module V2
  class MembershipsController < V2::ApiBaseController
    skip_before_action :authenticate, only: [:activate]
    load_and_authorize_resource :membership, :organisation, :user, except: [:activate]

    def index
      render json: MembershipSerializer.new(
        @memberships.select do |m|
          m.user == current_actor.user
        end, {
          params: { current_ability: current_ability },
          include: [:organisation, 'organisation.subscription']
        }
      ).serializable_hash.to_json
    end

    def create
      return render json: { error: 'missing parameters' } unless user_params[:email] && membership_params[:organisation_id]

      @organisation = Organisation.find(membership_params[:organisation_id])
      response = OrganisationService.invite_user_to_organisation(user_params[:email], @organisation.id, membership_params[:role], user_params[:name])

      render json: response
    end

    def activate
      membership = Membership.find_by(jwt_token_key: params[:id])
      raise UnauthorizedError unless membership

      if !membership.organisation.subscription.present? && membership.role_owner?
        OrganisationService.create_subscription(membership)
        membership.organisation.active!
      end

      membership.active!
      membership.user.activate! unless membership.user.activation_state == 'active'

      current_ability = Ability.new(Actor.new(:user, user: membership.user))

      base = Settings.external_frontend_url
      session = Session.new
      session.user = membership.user
      session.memberships = membership.user.memberships
      session.default_organisation = membership.organisation.public_id

      session.next_path = if !membership.user.has_seen_intro
                            "#{base}/dashboard/welcome"
                          else
                            "#{base}/dashboard"
                          end
      render json: SessionSerializer.new(session, { params: { current_ability: current_ability } })
    end

    def show
      render json: MembershipSerializer.new(@membership, {
        params: { current_ability: current_ability },
        include: [:organisation, 'organisation.subscription']
      }).serializable_hash.to_json
    end

    def update
      @membership.update(membership_params)

      render json: MembershipSerializer.new(
        @membership, {
          params: { current_ability: current_ability },
          include: [:organisation, 'organisation.subscription']
        }).serializable_hash.to_json
    end

    def destroy
      EventService.new.revoke_token([@membership])

      if @membership.organisation.memberships.count == 1
        OrganisationMailer.deleted_email(@membership.organisation).deliver_now
        OrganisationService.delete(organisation: @membership.organisation, soft_delete: true)
      end

      if @membership.user.memberships.count == 1
        UserMailer.deleted_email(@membership.user).deliver_now
        @membership.user.destroy
      end

      @membership.delete

      render json: { success: !@membership.persisted? }
    end

    private

    def membership_params
      params.require(:membership).permit(:organisation_id, :role, :notify_device_update, :notify_invoice)
    end

    def user_params
      params.require(:membership).permit(:email, :name)
    end
  end
end
