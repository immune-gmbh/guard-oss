require "test_helper"

class OauthsControllerTest < ActionDispatch::IntegrationTest
  test "should get oauth" do
    skip
    get oauths_oauth_url
    assert_response :success
  end

  test "should get callback" do
    skip
    get oauths_callback_url
    assert_response :success
  end

  test "should honor disable_registration" do
    skip
    Settings.authentication.disable_registration = true
    get "/v2/oauth/callback/google", params: { code: "abcd", client_id: "aa" }
    assert_response 403
  end
end
