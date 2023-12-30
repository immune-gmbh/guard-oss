class AlertMailer < ApplicationMailer
  helper MailerHelper

  before_action { @user = params.fetch(:membership).user }
  before_action { @organisation = params.fetch(:membership).organisation }
  before_action { @alerts = params.fetch(:alerts) }

  def alerts_email
    @by_device = {}
    @alerts.each do |a|
      @by_device[a.device] ||= []
      @by_device[a.device] << a
    end
    @earliest = @alerts.select {|a| a.timestamp != nil}.map {|a| a.timestamp}.sort.first

    if @by_device.entries.size > 1
      @subject = "#{@by_device.entries.size} devices manipulated"
    elsif @alerts.size > 1
      @subject = "#{@alerts.first.name} manipulated"
    else
      @subject = "#{@alerts.first.name}: #{@alerts.first.title}"
    end

    mail(
      to: @user.email,
      subject: @subject,
    )
  end
end
