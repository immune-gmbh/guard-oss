namespace :db do
  desc 'Grants access to all sequences and tables for a given role'
  task :chown, [:role] => :environment do |task, args|
    ActiveRecord::Base.transaction do 
      ActiveRecord::Base.connection.execute("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO #{args[:role].first}")
      ActiveRecord::Base.connection.execute("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO #{args[:role].first}")
    end
  end
end
