class AddAddressesToUsersAndOrganisationsAndSubscriptions < ActiveRecord::Migration[6.1]
  def change
    add_reference :users, :address, type: :uuid
    add_reference :organisations, :address, type: :uuid
    add_reference :subscriptions, :address, type: :uuid
  end
end
