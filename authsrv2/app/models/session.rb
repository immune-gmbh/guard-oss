class Session
  attr_accessor :default_organisation, :next_path, :user, :memberships

  def user_id
    @user.public_id
  end

  def membership_ids
    @memberships.map(&:public_id)
  end

  def id
    "1"
  end
end
