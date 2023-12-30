require "minitest/autorun"
require "test_helper"

class ProbesControllerTest < ActionDispatch::IntegrationTest
  test "full ready" do
    assert_emails 1 do
      get "/v2/ready", params: { full: true }
    end
    assert_response :ok
  end
end
