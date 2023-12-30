module V2
  class OrganisationsController < V2::ApiBaseController
    load_and_authorize_resource :organisation, expect: [:index]

    def index
      organisations = Organisation.where.not(status: :deleted).order(updated_at: :desc).select do |org|
        can? :read, org
      end

      render json: OrganisationSerializer.new(organisations, {include: [:subscription, :memberships], params: {current_ability: current_ability}}).serializable_hash.to_json
    end

    def create
      if params[:admin_view]
        response = OrganisationService.create_organisation_for_invite(
          organisation_params.merge(organisation_admin_params))
      else
        response = OrganisationService.create_organisation_and_subscription(
          current_actor.user,
          organisation_params.merge({ address: address_params }))
      end

      p response.organisation.address
      if response.success
        # return the membership only if the creator is also a member.
        render json: OrganisationSerializer.new(response.organisation, {
          include: ["memberships", "address"],
          fields: {
            membership: Membership.attribute_names + ["token","enrollment_token","name","user_name","user_email","can_delete","organisation_id"],
          },
          params: {current_ability: current_ability}
        }).serializable_hash
      else
        render status: :unprocessable_entity, json: {
          errors: [{id: "failed", title: response.message }]
        }
      end
    end

    def show
      include = conditional_include({
        "address": true,
        ["memberships"] => params["include"]&.include?("users"),
        ["users"] => params["include"]&.include?("users"),
      })
      fields = {
        membership: MembershipSerializer.attribute_names - [:organisation],
        user: UserSerializer.attribute_names - [:organisations]
      }

      render json: OrganisationSerializer.new(@organisation, {
        include: include,
        fields: fields,
        params: {current_ability: current_ability}
      }).serializable_hash.to_json
    end

    def update
      @errors = []
      new_address = false
      begin
        ActiveRecord::Base.transaction do
          # update adress if present
          if address_params && address_params.present?
            unless @organisation.address
              @organisation.build_address
              new_address = true
            end
            @organisation.address.update(address_params)
          end

          # update organisation fields in DB
          organisation_update_params = organisation_params.merge(organisation_admin_params)
          if @organisation.update(organisation_update_params)
            # update VAT ID on Stripe side
            if @organisation.vat_number_previously_changed?
              vat_response = OrganisationService.update_vat_number_with_stripe(@organisation)
              if !vat_response.success
                @errors << { id: "vat_number", title: vat_response.message }
                raise ActiveRecord::Rollback
              end
            end

            # update billing address on stripe side
            if @organisation.address&.previous_changes&.any?
              address_response = OrganisationService.update_address_and_taxes_with_stripe(@organisation)
              if !address_response.success
                @errors << { id: "address", title: address_response.message }
                raise ActiveRecord::Rollback
              end
            end

            # update device quota
            if (quota = organisation_virtual_params[:device_quota])
              quota = Integer(quota)
              if quota == nil || quota < 0
                @errors << { id: "device_quota", title: "less than zero" }
                raise ActiveRecord::Rollback
              else
                if !EventService.new.update_quota(@organisation, devices: quota)
                  # XXX: 501 is more approriate here
                  @errors << { id: "device_quota", title: "sending event failed" }
                  raise ActiveRecord::Rollback
                end
              end
            end
          else
            serialize_errors(@organisation)
            raise ActiveRecord::Rollback
          end
        end
      rescue => error
        trace_exception error
        @errors << { title: "Something went wrong" }
        # destroy the new address if we run in an error
        @organisation.address.destroy if new_address
      end

      # send jsonapi response
      if @errors.empty?
        render json: OrganisationSerializer.new(@organisation, {
          include: [:address],
          params: {current_ability: current_ability}
        }).serializable_hash
      else
        render status: :unprocessable_entity, json: { errors: @errors }
      end
    end

    def destroy
      # we should destroy all memberships and set the status to "deleted",
      # cancel the subscription on the spot and, stop issuing tokens,
      OrganisationService.delete(organisation: @organisation, soft_delete: true)

      if @organisation.deleted?
        render json: {}
      else
        render status: :bad_request, json: { errors: [{ id: 'failed' }] }
      end
    end

    private

    def address_params
      params.fetch(:address, {}).permit(:street_and_number, :postal_code, :city, :country)
    end

    def organisation_params
      params.permit(:name, :user_id, :vat_number, :splunk_enabled, :splunk_accept_all_server_certificates, :splunk_event_collector_url, :splunk_authentication_token, :syslog_enabled, :syslog_hostname_or_address, :syslog_udp_port, :invoice_name)
    end

    def organisation_admin_params
      return {} unless current_actor.user? && current_actor.user.admin?

      params.permit(:status, :freeloader)
    end

    def organisation_virtual_params
      return {} unless current_actor.user? && current_actor.user.admin?

      params.permit(:device_quota)
    end
  end
end
