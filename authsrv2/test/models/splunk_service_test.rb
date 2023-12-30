require "test_helper"
require "helpers/alerts_helper"

class SplunkServiceTest < ActiveSupport::TestCase
  include AlertsHelper

  test "formats failed appraisals" do
    a = mock_alert
    conn = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/splunk/services/collector" do |env|
          assert env.url.to_s == "https://siem.example.com/splunk/services/collector"
          assert env.request_headers["Authorization"] == "Splunk deadbeef"
          assert env.request_headers["Content-Type"] == "application/json"
          body = JSON.load env.body
          assert body["sourcetype"] == "immune:guard:json"
          assert body["event"]["signature_id"] == "pcr-changed"
          [200, {}, ""]
        end
      end
    end

    org = organisations(:kais_org)
    org.splunk_event_collector_url = "https://siem.example.com/splunk/services/collector"
    org.splunk_authentication_token = "deadbeef"

    SplunkService.new().send org, [a], conn
  end

  test "handles server errors" do
    a = mock_alert
    conn = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/splunk/services/collector" do |env|
          [500, {}, "500"]
        end
      end
    end

    org = organisations(:kais_org)
    org.splunk_event_collector_url = "https://siem.example.com/splunk/services/collector"
    org.splunk_authentication_token = "deadbeef"

    assert_raise Faraday::ServerError do
      SplunkService.new().send org, [a,a,a,a,a], conn
    end
  end

  test "handles client errors" do
    a = mock_alert
    conn = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/splunk/services/collector" do |env|
          [400, {}, "400"]
        end
      end
    end

    org = organisations(:kais_org)
    org.splunk_event_collector_url = "https://siem.example.com/splunk/services/collector"
    org.splunk_authentication_token = "deadbeef"

    assert_raise Faraday::ClientError do
      SplunkService.new().send org, [a,a,a,a], conn
    end
  end

  test "handles redirects" do
    a = mock_alert
    conn = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/splunk/services/collector" do |env|
          [303, {Location: "https://siem.example.com/redirect/splunk/services/collector"}, "303"]
        end
        stub.get "/redirect/splunk/services/collector" do |env|
          [200,{},"OK"]
        end
      end
    end

    org = organisations(:kais_org)
    org.splunk_event_collector_url = "https://siem.example.com/splunk/services/collector"
    org.splunk_authentication_token = "deadbeef"

    SplunkService.new().send org, [a,a,a,a,a], conn
  end
end
