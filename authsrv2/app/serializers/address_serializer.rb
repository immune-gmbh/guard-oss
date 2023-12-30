class AddressSerializer < BaseSerializer
  attributes :street_and_number, :city, :postal_code, :country
end
