require 'cef'

class SyslogService
  def send(org, alerts, client = nil)
    client ||= new_client org.syslog_hostname_or_address, org.syslog_udp_port
    alerts.each do |a|
      send_single client, a
    end
  end

  private

  def send_single(client, alert)
    # eats and records all exceptions
    GlobalTracer.in_span 'SyslogService::send' do |sp|
      sp.set_attribute 'code.file', __FILE__
      sp.set_attribute 'code.line', __LINE__
      sp.set_attribute 'alert', alert.inspect

      ev = CEF::Event.new(
        name: alert.title,
        deviceEventClassId: alert.id,
        deviceSeverity: '8',
        deviceHostName: alert.hostname,
        message: alert.description,
        receiptTime: alert.timestamp.to_s,
        deviceExternalId: alert.device,
        deviceCustomNumber1: alert.appraisal,
        deviceCustomNumber2: alert.name
      )

      ev.my_hostname = 'api.immu.ne'

      sp.set_attribute 'cef-event', ev.inspect
      client.emit ev
    end
  end

  def new_client(host, port)
    raise ArgumentError 'invalid hostname' if host&.to_s&.empty?

    port = 514 if port&.to_i&.zero?
    client = CEF::UDPSender.new host&.to_s, port&.to_i
    client.eventDefaults = {
      deviceProduct: 'immune Guard',
      deviceVendor: 'immune GmbH',
      deviceVersion: '2'
    }
    client
  end
end
