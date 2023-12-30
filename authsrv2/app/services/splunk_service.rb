class SplunkService
  def send(org, alerts, conn = nil)
    endpoint = org.splunk_event_collector_url

    GlobalTracer.in_span(
      'SplunkService::alert',
      attributes: {
        'http.url' => endpoint.to_s, 'http.method' => 'PUT', 'alerts' => alerts.inspect,
        'code.file' => __FILE__, 'code.line' => __LINE__
      },
      kind: :client
    ) do
      if conn
        do_send(org, alerts, conn)
      else
        Faraday.new do |my_conn|
          do_send(org, alerts, my_conn)
        end
      end
    end
  end

  private

  def do_send(org, alerts, conn)
    conn.response :logger
    conn.response :raise_error
    conn.use FaradayMiddleware::FollowRedirects, limit: 5
    conn.request :authorization, :Splunk, org.splunk_authentication_token
    conn.headers['Content-Type'] = 'application/json'
    conn.ssl[:verify] = !org.splunk_accept_all_server_certificates

    endpoint = org.splunk_event_collector_url
    alerts.each do |a|
      conn.post endpoint, JSON.dump((format a))
    end
  end

  def format(alert)
    {
      time: "#{alert.timestamp.to_i / 1000}.#{alert.timestamp.to_i % 1000}",
      host: 'api.immu.ne',
      source: 'immune Guard',
      sourcetype: 'immune:guard:json',
      event: {
        app: 'immune Guard',
        description: alert.description,
        dest: alert.name,
        dest_host: alert.hostname,
        id: alert.appraisal,
        mitre_technique_id: 'T1495',
        severity: 'high',
        signature: alert.title,
        signature_id: alert.id,
        type: 'alert'
      }
    }
  end
end
