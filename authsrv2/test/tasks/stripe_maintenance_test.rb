require "minitest/autorun"
require "test_helper"
require "rake"

class StripeMaintenance < ActionDispatch::IntegrationTest
  setup do
    Rake.application.rake_require "tasks/stripe_maintenance"
    Rake::Task.define_task(:environment)
  end

  test "finds missing customers" do
    customers = [
      OpenStruct.new({id: "cus_2"}),
      OpenStruct.new({id: "cus_3"}),
      OpenStruct.new({id: "cus_4"}),
    ]

    Stripe::Customer.stub :list, customers do
      assert_output /Customers missing from the database.\["cus_4"\].Customers deleted from Stripe\n\[".+?"\]/m do
        Rake.application.invoke_task "maintenance:stripe"
      end
    end
  end
end
