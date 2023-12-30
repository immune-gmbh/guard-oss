class EventService
  def initialize(client: nil)
    @endpoint = "#{Settings.internal_apisrv_url}/v2/events"
    @client = client || Faraday.new

    register_metrics
  end

  def revoke_token(memberships)
    lifetime = Settings.authentication.api_token_lifetime_seconds.to_i || 3600
    data = {
      expires: (Time.now + lifetime.seconds).to_datetime.rfc3339,
      token_ids: memberships.map(&:jwt_token_key)
    }
    send(nil, 'ne.immu.v2.revoke-credentials', data)
  end

  def update_quota(organisation, devices: 0)
    data = {
      devices: devices.to_s,
      features: ['attestation']
    }

    send(organisation, 'ne.immu.v2.quote-update', data)
  end

  private

  def register_metrics
    prometheus = Prometheus::Client.registry

    @events_outgoing_total = prometheus.get('events_outgoing_total')
    unless @events_outgoing_total.present?
      @events_outgoing_total =
        Prometheus::Client::Counter.new :events_outgoing_total,
                                        docstring: 'Counter of all outgoing cloudevents, partitioned by type, target and status',
                                        labels: %i[result endpoint type]
      prometheus.register @events_outgoing_total
    end

    @events_outgoing_latency = prometheus.get('events_outgoing_latency')
    unless @events_outgoing_latency.present?
      @events_outgoing_latency =
        Prometheus::Client::Histogram.new :events_outgoing_latency,
                                          docstring: 'Histogram of in flight cloudevent latency, partitioned by type, target and status',
                                          labels: %i[result endpoint type]
      prometheus.register @events_outgoing_latency
    end
  end

  def send(subject, type, body)
    return if Settings.development.disable_events

    ret = 'unknown'
    time = Benchmark.realtime do
      event =
        CloudEvents::Event.create spec_version: '1.0',
                                  id: SecureRandom.alphanumeric(16),
                                  source: Settings.service_name,
                                  type: type,
                                  subject: subject&.public_id,
                                  time: Time.now,
                                  data_content_type: 'application/json',
                                  data: body
      headers, body = CloudEvents::HttpBinding.default.encode_event event
      jwt = TokenService.issue_service_token(subject)

      GlobalTracer.in_span(
        'PUT /v2/events',
        attributes: { 'http.url' => @endpoint, 'http.method' => 'POST' },
        kind: :client
      ) do |span|
        OpenTelemetry.propagation.inject headers

        @client.response :logger
        @client.response :raise_error
        @client.request :authorization, :Bearer, jwt

        resp = @client.post @endpoint, body, headers
        span.set_attribute('http.status_code', resp.status)
        ret = resp.success? ? 'ack' : 'nack'
      rescue Faraday::ServerError, Faraday::ConnectionFailed
        ret = 'undelivered'
      rescue Faraday::ClientError
        ret = 'nack'
      end
    end

    @events_outgoing_latency.observe time, labels: { result: ret, endpoint: @endpoint, type: type }
    @events_outgoing_total.increment by: 1, labels: { result: ret, endpoint: @endpoint, type: type }

    ret == 'ack'
  end
end
