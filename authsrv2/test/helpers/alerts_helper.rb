require "minitest/autorun"

module AlertsHelper
  def mock_alert
    appraisal = {
      id: "1234",
      received: DateTime.new,
    }
    device = {
      id: "1234",
      name: "Device #1",
    }
    a = Alert.new(OpenStruct.new(device), OpenStruct.new(appraisal))
    a.id = "pcr-changed"
    a.title = "Firmware manipulated"
    a.appraisal = "42"
    a.timestamp = Time.now.to_i * 1000
    a.device = "42"
    a.name = "Test Device 1"
    a.description = "The Initial boot block was changed. This indicates an unauthorized reconfiguration or malware."
    a 
  end

  def mock_udpsocket
    double = Minitest::Mock.new
    def double.connect(host, port)
      raise ArgumentError.new "" if host != "siem.example.com"
      raise ArgumentError.new "" if port != 514
      true
    end
    def double.send(data, x)
      # CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
      raise ArgumentError.new "" if !(data =~ /<\d+>\w+ \d+ \d+ \d+:\d+:\d+ api.immu.ne CEF:0|immune GmbH|immune Guard|1|failed-appraisal|Failed appraisal|8|.+/)
      data.size
    end
    def double.nil?
      false
    end
    double
  end
end
