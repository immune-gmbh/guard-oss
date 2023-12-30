class UserMailer < ApplicationMailer
  def activation_needed_email(user)
    @user = user
    @url = "#{Settings.external_frontend_url}/registration/confirm?activationToken=#{@user.activation_token}"

    mail(to: @user.email, subject: 'Confirm your email address')
  end

  def invitation_email(membership)
    @user = membership.user
    @membership = membership

    @url = "#{Settings.external_frontend_url}/registration/confirm?membershipToken=#{@membership.jwt_token_key}"

    mail(to: @user.email, subject: "Invitation from immune Guard to join #{@membership.organisation.name}")
  end

  def reset_password_email(user)
    @user = user
    @url = "#{Settings.external_frontend_url}/set_new_password?resetToken=#{@user.reset_password_token}"
    mail(to: @user.email, subject: 'Your password reset link')
  end

  def deleted_email(user)
    @user = user
    mail(to: @user.email, subject: 'Your account has been deleted.')
  end
end
