class RemoveUserFromSubscription < ActiveRecord::Migration[6.1]
  def change
    remove_reference :subscriptions, :user
    remove_column :users, :stripe_customer_id
  end
end
