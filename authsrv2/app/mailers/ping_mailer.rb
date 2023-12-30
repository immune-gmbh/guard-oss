class PingMailer < ApplicationMailer
  def test_email
    mail(to: "test@example.com", subject: "Test eMail").tap do |msg|
      msg.mailgun_options = {
        "testmode" => true
      }
    end
  end
end
PingMailer.raise_delivery_errors = true
