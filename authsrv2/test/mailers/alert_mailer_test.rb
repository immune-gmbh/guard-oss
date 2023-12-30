require "test_helper"

class AlertMailerTest < ActionMailer::TestCase
  def mock_alert
    msg = OpenStruct.new(JSON.load(file_fixture('failed-appraisal.json').read))
    msg.data["device"]["id"] = SecureRandom.base58(5)
    msg.data["device"]["name"] = "Device #{msg.data["device"]["id"]}"
    Alert.failed_appraisal msg.data["device"], Appraisal.new(JSON.dump(msg.data["next"]))
  end

  def mock_alert_v2
    appr = Appraisal.new(file_fixture('no-secureboot.appraisal.json').read)
    msg = OpenStruct.new(JSON.load(file_fixture('failed-appraisal.json').read))
    msg.data["device"]["id"] = SecureRandom.base58(5)
    msg.data["device"]["name"] = "Device #{msg.data["device"]["id"]}"
    Alert.failed_appraisal msg.data["device"], appr
  end

  setup do
    @membership = memberships(:kai_at_kais_org)
  end

  test "single device, multiple alerts" do
    @alert = mock_alert

    email = AlertMailer.with(membership: @membership, alerts: @alert).alerts_email

    assert_emails 1 do
      email.deliver_now
    end

    assert_equal ["support@immune.gmbh"], email.from
    assert_equal ["kai.michaelis@immu.ne"], email.to
    assert_match /Device .....: Firmware manipulated/, email.subject
  end

  test "single device, single alert" do
    @alert = mock_alert

    email = AlertMailer.with(membership: @membership, alerts: [@alert.first]).alerts_email

    assert_match /Device .....: .+/, email.subject
  end

  test "single device, single alert, v2" do
    @alert = mock_alert_v2

    email = AlertMailer.with(membership: @membership, alerts: [@alert.first]).alerts_email

    assert_match /Device .....: Device configuration.+/, email.subject
  end


  test "5 devices, multiple alerts" do
    @alerts = 5.times.map do mock_alert end.to_a.flatten

    email =  AlertMailer.with(membership: @membership, alerts: @alerts).alerts_email

    assert_equal "5 devices manipulated", email.subject
  end
end
