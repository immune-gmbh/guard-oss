namespace :db do
  desc 'Waits for the database to become available'
  task wait: :environment do
    loop do
      ActiveRecord::Base.connection.execute("SELECT 1")
      break
    rescue ActiveRecord::ConnectionNotEstablished, ActiveRecord::NoDatabaseError => e
      puts e
      sleep 2.seconds
    end
  end
end
