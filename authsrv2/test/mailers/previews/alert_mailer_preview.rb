# Preview all emails at http://localhost:3000/rails/mailers/alert_mailer
class AlertMailerPreview < ActionMailer::Preview
  def mock_alert
    msg = OpenStruct.new(JSON.load(File.open('test/fixtures/files/failed-appraisal-v5.json').read))
    msg.data["device"]["id"] = SecureRandom.base58(5)
    msg.data["device"]["name"] = "Device #{msg.data["device"]["id"]}"
    Alert.failed_appraisal msg.data["device"], Appraisal.new(JSON.dump(msg.data["next"]))
  end

  def single_device_multiple_alerts_email
    @membership = Membership.first
    @alert = mock_alert

    AlertMailer.with(membership: @membership, alerts: @alert).alerts_email
  end

  def single_device_single_alert_email
    @membership = Membership.first
    @alert = mock_alert

    AlertMailer.with(membership: @membership, alerts: [@alert.first]).alerts_email
  end

  def multiple_devices_multiple_alerts_email
    @membership = Membership.first
    @alerts = 3.times.map do mock_alert end.flatten

    AlertMailer.with(membership: @membership, alerts: @alerts).alerts_email
  end
end
