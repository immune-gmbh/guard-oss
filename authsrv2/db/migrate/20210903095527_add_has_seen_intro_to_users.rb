class AddHasSeenIntroToUsers < ActiveRecord::Migration[6.1]
  def change
    add_column :users, :has_seen_intro, :boolean, default: false
  end
end
