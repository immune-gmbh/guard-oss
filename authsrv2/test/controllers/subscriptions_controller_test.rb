require "minitest/autorun"
require "test_helper"

class SubscriptionsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @org = organisations(:kais_org)
    @membership = memberships(:kai_at_kais_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    @svc_err = OpenStruct.new({ success: false, message: "blah" })
  end

  test "setup intent" do
    intent = OpenStruct.new(JSON.load(file_fixture("setup_intent.json").read))
    @svc_ok = OpenStruct.new({
      success: true,
      subscription: @org.subscription,
      setup_intent: intent
    })
    OrganisationService.stub :setup_payment, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/intent", params: { organisation_id: @org.id }, headers: @headers
        assert_response :ok
        integration_test("post-subscriptions-intent.json")
      end
    end
    assert_equal JSON.load(@response.body).dig("data", "attributes", "setup_intent_secret"), intent.client_secret
  end

  test "setup intent -- payment disabled" do
    Settings.payment.disable = true

    post "/v2/subscriptions/intent", params: { organisation_id: @org.id }, headers: @headers
    assert_response :service_unavailable

    Settings.payment.disable = false
  end

  test "setup intent unauthorized" do
    @org = organisations(:testys_org)
    intent = OpenStruct.new(JSON.load(file_fixture("setup_intent.json").read))
    @svc_ok = OpenStruct.new({
      success: true,
      subscription: @org.subscription,
      setup_intent: intent
    })
    OrganisationService.stub :setup_payment, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/intent", params: { organisation_id: @org.id }, headers: @headers
        assert_response :forbidden
      end
    end
  end

  test "unsuccessful setup intent" do
    OrganisationService.stub :setup_payment, @svc_err do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/intent", params: { organisation_id: @org.id }, headers: @headers
        assert_response :service_unavailable
      end
    end
  end

  test "set default payment" do
    @svc_ok = OpenStruct.new({ success: true, subscription: @org.subscription })
    req = { organisation_id: @org.id, payment_method_id: "test" }
    OrganisationService.stub :make_payment_method_default, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/default_payment_method", params: req, headers: @headers
        assert_response :ok
        assert_equal JSON.load(@response.body).pretty_inspect,
          SubscriptionSerializer.new(@org.subscription).as_json.pretty_inspect
        integration_test("post-subscriptions-default_payment_method.json")
      end
    end
  end

  test "set default payment -- payment disabled" do
    Settings.payment.disable = true

    req = { organisation_id: @org.id, payment_method_id: "test" }
    post "/v2/subscriptions/default_payment_method", params: req, headers: @headers
    assert_response :service_unavailable

    Settings.payment.disable = false
  end

  test "unsuccessful default payment change" do
    req = { organisation_id: @org.id, payment_method_id: "test" }
    OrganisationService.stub :make_payment_method_default, @svc_err do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/default_payment_method", params: req, headers: @headers
        assert_response :service_unavailable
      end
    end
  end

  test "set default payment unauthorized" do
    @org = organisations(:testys_org)
    @svc_ok = OpenStruct.new({ success: true, subscription: @org.subscription })
    req = { organisation_id: @org.id, payment_method_id: "test" }
    OrganisationService.stub :make_payment_method_default, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions/default_payment_method", params: req, headers: @headers
        assert_response :forbidden
      end
    end
  end

  test "create subscription" do
    @svc_ok = OpenStruct.new({ success: true, subscription: @org.subscription })
    OrganisationService.stub :create_subscription, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions", params: { membership_id: @membership.id }, headers: @headers
        assert_response :ok
        assert_equal JSON.load(@response.body).pretty_inspect,
          SubscriptionSerializer.new(@org.subscription).as_json.pretty_inspect
      end
    end
  end

  test "create subscription -- payment disabled" do
    Settings.payment.disable = true
    @org.update!(stripe_customer_id: nil, subscription: nil)
    post "/v2/subscriptions", params: { membership_id: @membership.id }, headers: @headers
    assert_response :service_unavailable
    Settings.payment.disable = false
  end

  test "failed create subscription" do
    OrganisationService.stub :create_subscription, @svc_err do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions", params: { membership_id: @membership.id }, headers: @headers
        assert_response :service_unavailable
      end
    end
  end

  test "unauthorized create subscription" do
    @membership = memberships(:kai_at_testys_org)
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    @svc_ok = OpenStruct.new({ success: true, subscription: @org.subscription })
    OrganisationService.stub :create_subscription, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions", params: { membership_id: @membership.id }, headers: @headers
        assert_response :forbidden
      end
    end
  end

  test "create subscription for inexistent membership" do
    @svc_ok = OpenStruct.new({ success: true, subscription: @org.subscription })
    OrganisationService.stub :create_subscription, @svc_ok do 
      Stripe::Customer.stub :retrieve, "id" do
        post "/v2/subscriptions", params: { membership_id: "blah" }, headers: @headers
        assert_response :not_found
      end
    end
  end

  test "get subscription" do
    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/subscriptions/#{@org.subscription.id}", headers: @headers
      assert_response :ok
      assert_equal JSON.load(@response.body).pretty_inspect,
        SubscriptionSerializer.new(@org.subscription).as_json.pretty_inspect
    end
  end

  test "list subscriptions" do
    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/subscriptions", headers: @headers
      assert_response :ok
      assert_equal JSON.load(@response.body)["data"].size, 1
    end
  end
end
