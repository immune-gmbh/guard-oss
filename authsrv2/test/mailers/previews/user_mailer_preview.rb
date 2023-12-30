class UserMailerPreview < ActionMailer::Preview
  def invitation_email
    @user = User.first
    @membership = Membership.first

    UserMailer.with(user: @user).invitation_email(@membership)
  end
end