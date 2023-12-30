require "test_helper"
require "helpers/alerts_helper"

class SyslogServiceTest < ActiveSupport::TestCase
  include AlertsHelper

  test "formats failed appraisals" do
    a = mock_alert
    org = organisations(:kais_org)
    org.syslog_hostname_or_address = "siem.example.com"
    org.syslog_udp_port = 514

    UDPSocket.stub :new, mock_udpsocket do
      SyslogService.new().send org, [a,a,a,a,a]
    end
  end

  test "handles ill formed port" do
    a = mock_alert
    org = organisations(:kais_org)
    org.syslog_hostname_or_address = "siem.example.com"
    org.syslog_udp_port = "aa"

    UDPSocket.stub :new, mock_udpsocket do
      SyslogService.new().send org, [a,a,a,a,a]
    end
  end
end
