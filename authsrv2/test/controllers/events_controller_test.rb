require 'minitest/autorun'
require 'test_helper'
require 'helpers/events_helper'

class EventsControllerTest < ActionDispatch::IntegrationTest
  include EventsHelper

  test 'should handle unknown event' do
    token = TokenService.issue_service_token organisations(:kais_org)
    headers, body = unknown_event
    headers['Authorization'] = "Bearer #{token}"

    post '/v2/events', params: body, headers: headers
    assert_response :bad_request
  end

  test 'should handle wrong subjects' do
    headers, body = new_appraisal subject: 'non-existent', verdict: false

    post '/v2/events', params: body, headers: headers
    assert_response :unauthorized
  end

  test 'should authenticate requests' do
    headers, body = new_appraisal verdict: false

    post '/v2/events', params: body, headers: headers
    assert_response :unauthorized
  end

  test 'should process failed appraisals' do
    token = TokenService.issue_service_token organisations(:kais_org)
    fixtures = [
      'failed-appraisal.json',
      'failed-appraisal-v2.json',
      'failed-appraisal-v3.json',
      'failed-appraisal-v4.json',
      'failed-appraisal-v5.json',
      'failed-appraisal-v6.json',
    ]

    fixtures.each do |f|
      body = JSON.parse(file_fixture(f).read)
      body['subject'] = organisations(:kais_org).public_id
      headers = {
        'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
        'Authorization' => "Bearer #{token}"
      }
      syslog_mock = Minitest::Mock.new
      syslog_mock.expect(:send, true) do |org, alerts|
        org.public_id == organisations(:kais_org).public_id && !alerts.empty?
      end
      splunk_mock = Minitest::Mock.new
      splunk_mock.expect(:send, true) do |org, alerts|
        org.public_id == organisations(:kais_org).public_id && !alerts.empty?
      end

      Settings.features[:alert_emails] = [organisations(:kais_org).public_id]

      assert_emails 1 do
        SyslogService.stub :new, syslog_mock do
          SplunkService.stub :new, splunk_mock do
            post '/v2/events', params: JSON.dump(body), headers: headers
            assert_response :accepted
          end
        end
      end

      assert_mock syslog_mock
      assert_mock splunk_mock
    end
  end

  test 'should accept successful appraisals' do
    token = TokenService.issue_service_token organisations(:fixed_id_org)
    body = JSON.parse(file_fixture('good-appraisal.json').read)
    body['subject'] = organisations(:fixed_id_org).public_id
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: JSON.dump(body), headers: headers
    assert_response :ok
  end

  test 'should accept failed appraisals for vulnerable devices' do
    token = TokenService.issue_service_token organisations(:fixed_id_org)
    body = JSON.parse(file_fixture('still-failing-appraisal.json').read)
    body['subject'] = organisations(:fixed_id_org).public_id
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: JSON.dump(body), headers: headers
    assert_response :ok
  end

  test 'should process expired appraisals' do
    token = TokenService.issue_service_token organisations(:kais_org)
    body = JSON.parse(file_fixture('expired-appraisal.json').read)
    body['subject'] = organisations(:kais_org).public_id
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    syslog_mock = Minitest::Mock.new
    syslog_mock.expect(:send, true) do |org, alerts|
      org.public_id == organisations(:kais_org).public_id && !alerts.empty?
    end
    splunk_mock = Minitest::Mock.new
    splunk_mock.expect(:send, true) do |org, alerts|
      org.public_id == organisations(:kais_org).public_id && !alerts.empty?
    end

    Settings.features[:alert_emails] = [organisations(:kais_org).public_id]

    assert_emails 1 do
      SyslogService.stub :new, syslog_mock do
        SplunkService.stub :new, splunk_mock do
          post '/v2/events', params: JSON.dump(body), headers: headers
          assert_response :accepted
        end
      end
    end

    assert_mock syslog_mock
    assert_mock splunk_mock
  end

  test 'should process heartbeats' do
    token = TokenService.issue_service_token organisations(:kais_org)
    headers, body = heartbeat update_usage_records: true
    headers['Authorization'] = "Bearer #{token}"

    post '/v2/events', params: body, headers: headers
    assert_response :ok
  end

  test 'should accept heartbeats w/o actions' do
    token = TokenService.issue_service_token organisations(:kais_org)
    headers, body = heartbeat update_usage_records: false
    headers['Authorization'] = "Bearer #{token}"

    post '/v2/events', params: body, headers: headers
    assert_response :ok
  end

  test 'should handle billing updates' do
    token = TokenService.issue_service_token
    body = file_fixture('billing-update.json').read
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: body, headers: headers
    assert_response :accepted
  end

  test 'should refuse invalid billing updates' do
    token = TokenService.issue_service_token
    body = file_fixture('invalid-billing-update.json').read
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: body, headers: headers
    assert_response :bad_request
  end

  test 'should refuse billing updates to non existent orgs' do
    token = TokenService.issue_service_token
    body = file_fixture('unknown-org-billing-update.json').read
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: body, headers: headers
    assert_response :bad_request
  end

  test 'should handle empty billing updates' do
    token = TokenService.issue_service_token
    body = file_fixture('empty-billing-update.json').read
    headers = {
      'Content-Type' => 'application/cloudevents+json; charset=UTF-8',
      'Authorization' => "Bearer #{token}"
    }

    post '/v2/events', params: body, headers: headers
    assert_response :ok
  end
end
