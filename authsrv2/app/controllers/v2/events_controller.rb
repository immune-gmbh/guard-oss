module V2
  class EventsController < V2::ApiBaseController
    rescue_from ActiveRecord::RecordNotFound, with: :error_bad_request
    rescue_from ArgumentError, with: :error_bad_request
    rescue_from CloudEvents::FormatSyntaxError, with: :error_bad_request
    rescue_from CloudEvents::SpecVersionError, with: :error_bad_request

    def initialize
      super
      @binding = CloudEvents::HttpBinding.default
    end

    def receive
      event = @binding.decode_event request.env
      @data = OpenStruct.new(event.data)
      @subject = event.subject

      current_span.set_attribute 'cloudevents.event_type', event.type
      current_span.set_attribute 'cloudevents.event_id', event.id

      case event.type
      when 'ne.immu.v2.new-appraisal'
        trace('V2::EventsController::handle_new_appraisal') { handle_new_appraisal }

      when 'ne.immu.v2.appraisal-expired'
        trace('V2::EventsController::handle_expired_appraisals') { handle_expired_appraisals }

      when 'ne.immu.v2.heartbeat'
        trace('V2::EventsController::handle_heartbeat') { handle_heartbeat }

      when 'ne.immu.v2.billing-update'
        trace('V2::EventsController::handle_billing_update') { handle_billing_update }

      else
        render status: :bad_request
      end
    end

    private

    def handle_new_appraisal
      load_and_authorize_memberships

      next_appr = Appraisal.new(JSON.dump(@data['next']))
      prev_appr = @data['previous'] && Appraisal.new(JSON.dump(@data['previous']))

      current_span.set_attribute 'immune.next-appraisal', next_appr.id
      current_span.set_attribute 'immune.previous-appraisal', prev_appr&.id if prev_appr
      current_span.set_attribute 'immune.device', @data['device']&.[]('id') if @data['device']
      current_span.set_attribute 'immune.device-name', @data['device']&.[]('name') if @data['device']

      trace_event "Next verdict: #{next_appr.verdict}, previous verdict: #{prev_appr&.verdict}"

      if !next_appr.verdict && (prev_appr.nil? || prev_appr.verdict)
        trace_event "Appraisal verdict is false, sending alerts to #{@memberships} users"
        @alerts = Alert.failed_appraisal @data.device, next_appr
        trace_event "Generated #{@alerts.size} alerts"
        send_alerts

        render status: :accepted
      else
        render status: :ok
      end
    end

    def handle_expired_appraisals
      load_and_authorize_memberships

      trace_event "Appraisal expired, sending alerts to #{@memberships} users"
      @alerts = Alert.expired_appraisal(@data.device, Appraisal.new(JSON.dump(@data.previous)))
      trace_event "Generated #{@alerts.size} alerts"
      send_alerts

      render status: :accepted
    end

    def handle_heartbeat
      render status: :ok
    end

    def handle_billing_update
      ActiveRecord::Base.transaction do
        @data['records'].to_a.each do |rec|
          orgid = rec['organisation']
          next if orgid == 'ext-1'

          org = Organisation.find_by! public_id: orgid.to_s
          authorize! :bill, org

          devs = rec['devices'].to_i
          raise ArgumentError if devs.negative?

          # limit device update notification to max once a day
          OrganisationMailer.device_usage_will_update_email(org).deliver_now unless org.subscription.new_devices_amount
          # cron job will create usage records for all org.subs.new_devices_amount != nil
          org.subscription.update!(new_devices_amount: devs)
        end
      end

      num_processed = @data['records'].to_a.size
      trace_event "Processed #{num_processed} billing records"

      if num_processed.positive?
        render status: :accepted
      else
        render status: :ok
      end
    end

    def send_alerts
      trace 'V2::EventsController::send_alerts' do
        return if @alerts.empty?

        @memberships.each do |m|
          next if !m.notify_device_update || !m.user.email

          AlertMailer.with(membership: m, alerts: @alerts).alerts_email.deliver_now
          trace_event "Scheduled mail to #{m.user.email}"
        end

        begin
          SplunkService.new.send(@organisation, @alerts) if @organisation.splunk_enabled
        rescue StandardError => e
          span = OpenTelemetry::Trace.current_span
          span.record_exception e
        end

        begin
          SyslogService.new.send(@organisation, @alerts) if @organisation.syslog_enabled
        rescue StandardError => e
          span = OpenTelemetry::Trace.current_span
          span.record_exception e
        end
      end
    end

    def load_and_authorize_memberships
      current_span.set_attribute 'cloudevents.event_subject', @subject

      if @subject == 'ext-1'
        @organisation = Organisation.first
        @memberships = []
      else
        @organisation = Organisation.find_by! public_id: @subject
        @memberships = if feature?(:alert_emails)
                         @organisation.memberships.active.where notify_device_update: true
                       else
                         []
                       end
        authorize! :alert, @organisation
      end
    end

    def error_bad_request
      render status: :bad_request
    end
  end
end
