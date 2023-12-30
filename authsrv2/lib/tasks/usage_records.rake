namespace :usage_records do
  desc 'Records daily usage for each subscription'
  task :create => :environment do
    subscriptions = Subscription.where(status: 'active').where.not(new_devices_amount: nil)

    today = Date.today
    subscriptions.each do |sub|
      ActiveRecord::Base.transaction do
        # Create the usage record for tracking.. well.. usage
        sub.usage_records.create(amount: sub.new_devices_amount, date: today)
        # Only send the mail if
        # - it is not the inital usage_record setup after registration
        # - the amount has changed
        if (sub.current_devices_amount != sub.new_devices_amount)
          OrganisationMailer.devices_updated_email(sub.organisation).deliver_now
        end
        # Reset usage_record trigger condition
        sub.update(current_devices_amount: sub.new_devices_amount, new_devices_amount: nil)
        # Transmit fresh usage projection to stripe to update the upcoming invoice amount
        StripeService.update_usage(sub)
      end
    end
  end
end
