namespace :maintenance do
  desc 'Fetches all customers and subscriptions from Stripe and compares them to what we have in the database.'
  task :stripe => :environment do
    customers = Hash.new
    new_customers = Stripe::Customer.list({limit: 100})

    for i in 0..100 do
      got_new = new_customers.reduce(false) do |acc, cus|
        acc |= customers[cus.id] == nil
        customers[cus.id] = cus
        acc
      end

      break unless got_new
      new_customers = Stripe::Customer.list({limit: 25, created: {lte: new_customers.last.created}})
    end

    at_stripe_but_not_here = customers.keys.to_set - Organisation.where(stripe_customer_id: customers.keys, status: "active").map(&:stripe_customer_id).to_set
    here_but_not_at_stripe = Organisation.where.not(stripe_customer_id: customers.keys).select {|org| org.active?}

    if !at_stripe_but_not_here.empty?
      puts "Customers missing from the database"
      p at_stripe_but_not_here.to_a
    end

    if !here_but_not_at_stripe.empty?
      puts "Customers deleted from Stripe"
      p here_but_not_at_stripe.map(&:public_id)
    end
  end
end
