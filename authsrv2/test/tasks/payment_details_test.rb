require "test_helper"
require "rake"

class PaymentDetailsTest < ActionDispatch::IntegrationTest
  setup do
    Rake.application.rake_require "tasks/payment_details"
    Rake::Task.define_task(:environment)
  end

  test "send reminder email" do
    assert_emails 1 do
      Rake.application.invoke_task "payment_details:remind"
    end
  end
end
