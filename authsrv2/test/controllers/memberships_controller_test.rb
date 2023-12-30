require "minitest/autorun"
require "test_helper"

class MembershipsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @org = organisations(:kais_org)
    @membership = memberships(:kai_at_kais_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @request_body = @membership.serializable_hash
   @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    @svc_ok = OpenStruct.new({ success: true })
    @svc_err = OpenStruct.new({ success: false, message: "blah" })
  end

  test "change email prefs" do
    Stripe::Customer.stub :retrieve, "id" do
      patch "/v2/memberships/#{@membership.id}", params: { membership: { notify_invoice: true, notify_device_update: true } }, headers: @headers
    assert_response :ok
    @membership.notify_device_update = true
    @membership.notify_invoice = true
    @response_body = MembershipSerializer.new(@membership, {
      params: {current_ability: @ability},
      include: [:organisation, "organisation.subscription"]
    }).as_json
    @response_body["data"]["attributes"]["enrollment_token"] = ""
    @response_body["data"]["attributes"]["token"] = ""
    resp = JSON.load(@response.body)
    resp["data"]["attributes"]["token"] = ""
    resp["data"]["attributes"]["enrollment_token"] = ""
    assert_equal resp.pretty_inspect, @response_body.pretty_inspect
    end
  end
end
