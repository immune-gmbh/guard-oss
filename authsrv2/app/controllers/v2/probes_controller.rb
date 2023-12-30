module V2
  class ProbesController < V2::ApiBaseController
    skip_before_action :authenticate

    def ready
      # check db connection
      ActiveRecord::Base.connection.execute 'SELECT 1'

      if params['full']
        p PingMailer.test_email.deliver_now
      end

      # check agent urls
      if Rails.env.production?
        Settings.agent_urls.each do |k,v|
          Faraday.new(v) do |conn|
            conn.response :raise_error
          end.head
        end
      end

      render plain: '', status: :ok
    rescue ActiveRecord::ConnectionNotEstablished => e
      trace_exception e
      render plain: '', status: :service_unavailable
    rescue Mailgun::CommunicationError => e
      trace_exception e
      render plain: '', status: :service_unavailable
    rescue Faraday::Error => e
      trace_exception e
      render plain: '', status: :service_unavailable
    end

    def healthy
      if KeyDiscoveryService.ping
        render plain: '', status: :ok
      else
        render plain: '', status: :service_unavailable
      end
    end
  end
end
