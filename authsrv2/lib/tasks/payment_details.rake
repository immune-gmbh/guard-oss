namespace :payment_details do
  desc 'Looks for 2 week old subscriptions asks for feedback and reminds them to provide payment info to use the service after the first month'
  task :remind => :environment do
    organisations = Organisation.where(status: 'active', created_at: 14.days.ago..13.days.ago)

    organisations.map do |org|
      OrganisationMailer.payment_reminder_email(org).deliver_now
    end
  end
end
