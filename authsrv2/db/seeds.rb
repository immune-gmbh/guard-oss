# This file should contain all the record creation needed to seed the database with its default values.
# The data can then be loaded with the bin/rails db:seed command (or created alongside the database with db:setup).
#
# Examples:
#
#   movies = Movie.create([{ name: 'Star Wars' }, { name: 'Lord of the Rings' }])
#   Character.create(name: 'Luke', movie: movies.first)
address = Address.create(street_and_number: "Kortumstraße 19-22", city: "Bochum", postal_code: "44789", country: "Germany")
User.create!(name: 'Tobe', email: 'tobe@example.com', password: 'secure', address: address)
User.create!(name: 'Marv', email: 'marv@example.com', password: 'secret', address: address)

User.all.each do |user|
  user.activate!
end

Organisation.create!(name: "#{User.first.name}'s Organisation", users: [User.second])
Organisation.create!(name: "#{User.second.name}'s Organisation", users: [User.first])

Membership.create(user: User.first, organisation: Organisation.first, role: :owner)
Membership.create(user: User.second, organisation: Organisation.second, role: :owner)


Subscription.create!(organisation: Organisation.first, status: 'active')
Subscription.create!(organisation: Organisation.last, status: 'active')
