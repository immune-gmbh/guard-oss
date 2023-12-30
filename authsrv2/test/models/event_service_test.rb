require "minitest/autorun"
require "test_helper"

class EventServiceTest < ActiveSupport::TestCase
  teardown do
    Faraday.default_connection = nil
  end

  test "we can send quota updated" do
    client = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/v2/events" do |env|
          assert env.request_headers["Content-Type"] == "application/json"
          assert env.url.to_s == "http://apisrv-v2.svc.default.cluster.local:8000/v2/events"
          assert env.method == :post
          body = JSON.load env.body
          puts env.body
          assert body["devices"] == "12"
          assert body["features"] == ["attestation"]
          [200, {}, ""]
        end
      end
    end
    svc = EventService.new client: client

    assert svc.update_quota(organisations(:kais_org), devices: 12)
  end

  test "we can revoke tokens" do
    client = Faraday.new do |b|
      b.adapter :test do |stub|
        stub.post "/v2/events" do |env|
          assert env.request_headers["Content-Type"] == "application/json"
          assert env.url.to_s == "http://apisrv-v2.svc.default.cluster.local:8000/v2/events"
          assert env.method == :post
          body = JSON.load env.body
          puts env.body
          assert body["token_ids"].size == users(:kai).memberships.size
          assert DateTime.rfc3339(body["expires"]) > DateTime.now
          [200, {}, ""]
        end
      end
    end
    svc = EventService.new client: client

    assert svc.revoke_token(users(:kai).memberships)
  end
end
