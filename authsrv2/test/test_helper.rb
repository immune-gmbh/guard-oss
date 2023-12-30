# code coverage info
require "simplecov"
require "simplecov-cobertura"

ENV['RAILS_ENV'] ||= 'test'
ENV['DISABLE_BOOTSNAP'] ||= 'true'
ENV['DISABLE_SPRING'] ||= 'true'
ENV['PARALLEL_WORKERS'] ||= '1'

SimpleCov.start "rails" do
  enable_coverage :branch

  add_filter "/test/"
  add_filter "/config/"

  add_group "Controllers", "app/controllers"
  add_group "Services", "app/services"
  add_group "Models", "app/models"
  add_group "Helpers", "app/helpers"
  add_group "Libraries", "lib"
end
SimpleCov.formatter = SimpleCov::Formatter::MultiFormatter.new([
  SimpleCov::Formatter::HTMLFormatter,
  SimpleCov::Formatter::CoberturaFormatter,
])

require_relative "../config/environment"
require "rails/test_help"

class ActiveSupport::TestCase
  # Run tests in parallel with specified workers
  parallelize(workers: :number_of_processors)

  # Setup all fixtures in test/fixtures/*.yml for all tests in alphabetical order.
  fixtures :all

  def integration_test(str)
    write_fixture(str, JSON.pretty_generate(JSON.load(@response.body)))
  end

  def write_fixture(filename, contents)
    File.open("test/fixtures/output/#{filename}", "w+") do |fd|
      fd.write(contents)
    end
  end

  # Add more helper methods to be used by all tests here...
end
