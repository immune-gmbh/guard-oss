require "minitest/autorun"
require "test_helper"

class OrganisationsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @org = organisations(:kais_org)
    @membership = memberships(:kai_at_kais_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @request_body = @org.serializable_hash
    @request_body["address"] = @org.address.serializable_hash
    @response_body = OrganisationSerializer.new(@org, {
      include: [:address],
      params: { current_ability: @ability }
    }).as_json
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    @svc_ok = OpenStruct.new({ success: true })
    @svc_err = OpenStruct.new({ success: false, message: "blah" })
  end

  test "create org -- issue #1006" do
    addr = {
      "country" => "Algeria",
      "city" => "a",
      "street_and_number" => "a",
      "postal_code" => "a"
    }
    org = {
      "name":"a",
      "invoice_name":"a",
      "vat_number":"",
    }
    org = OpenStruct.new(org.merge({ "id": "1", "address": OpenStruct.new(addr.merge({ "id": "1" })) }))
    org = OrganisationSerializer.new(org, { include: [ :address ] }).as_json
    sub = OpenStruct.new({
      id: 1,
      items: OpenStruct.new({ data: [OpenStruct.new({ id: 1 })] }),
      current_period_start: 1,
      current_period_end: 1,
    })
    Stripe::Subscription.stub :create, sub do
      Stripe::Customer.stub :create, OpenStruct.new({ id: 1 }) do
        post "/v2/organisations", params: org.merge({ "address": addr }), headers: @headers
        assert_response :ok
      end
    end
    resp = JSON.load(@response.body)
    assert_equal resp["included"][1]["attributes"], addr

    addr2 = Organisation.find(resp["data"]["id"]).address.as_json
    addr2.delete "public_id"
    addr2.delete "created_at"
    addr2.delete "updated_at"
    assert_equal addr2, addr
  end

  test "change nothing" do
    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :ok
    assert_equal JSON.load(@response.body), @response_body

    @request_body.delete "address"
    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :ok
  end

  test "invalid address change" do
    attr = {
      "street_and_number" => "",
    }
    @request_body["address"].merge! attr

    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :unprocessable_entity
      assert_equal JSON.load(@response.body)["errors"].size, 1
  end

  test "invalid org change" do
    attr = {
      "name" => organisations(:testys_org).name
    }
    @request_body.merge! attr

    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :unprocessable_entity
    assert_equal JSON.load(@response.body)["errors"].size, 1
  end

  test "should handle splunk update" do
    attr = {
      "splunk_enabled" => true,
      "splunk_event_collector_url" => 'http://siem.example.com/splunk/hec/test',
      "splunk_authentication_token" => 'deadbeefcafebabe',
      "splunk_accept_all_server_certificates" => false,
    }
    @request_body.merge! attr

    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :ok
    @response_body["data"]["attributes"].merge! attr
    assert_equal JSON.load(@response.body), @response_body
  end

  test "should handle vat update" do
    attr = {
      "vat_number" => "DE345789003",
    }
    @request_body.merge! attr

    OrganisationService.stub :update_vat_number_with_stripe, @svc_err do
      patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
      assert_response :unprocessable_entity
      assert_equal JSON.load(@response.body)["errors"].size, 1

      # db contents unchanged
      @org.reload
      assert_not_equal @org.vat_number, attr["vat_number"]
    end

    OrganisationService.stub :update_vat_number_with_stripe, @svc_ok do
      patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
      assert_response :ok

      # response ok
      @response_body["data"]["attributes"].merge! attr
      assert_equal JSON.load(@response.body).pretty_inspect, @response_body.pretty_inspect

      # db contents changed
      @org.reload
      assert_equal @org.vat_number, attr["vat_number"]
    end
  end

  test "should handle address update" do
    attr = {
      "street_and_number" => "test str. 11"
    }
    @request_body["address"].merge! attr

    OrganisationService.stub :update_address_and_taxes_with_stripe, @svc_err do
      patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
      assert_response :unprocessable_entity

      assert_equal JSON.load(@response.body)["errors"].size, 1

      @org.reload
      assert_not_equal @org.address.street_and_number, "test str. 11"
    end

    OrganisationService.stub :update_address_and_taxes_with_stripe, @svc_ok do
      patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
      assert_response :ok

      @org.reload
      assert @org.address.street_and_number, "test str. 11"

      @response_body["included"][0]["attributes"].merge! attr
      assert_equal JSON.load(@response.body).pretty_inspect, @response_body.pretty_inspect
    end
  end

  test "should handle address update -- payment disabled" do
    attr = {
      "street_and_number" => "test str. 11"
    }
    @request_body["address"].merge! attr

    Settings.payment.disable = true
    patch "/v2/organisations/#{@org.id}", params: @request_body, headers: @headers
    assert_response :ok

    @org.reload
    assert_equal @org.address.street_and_number, "test str. 11"
    Settings.payment.disable = false
  end

  test "gen test files" do
    #get "/v2/organisations", headers: @headers
    #assert_response :ok

    #File.open("get-organisations.json", "w+") do |fd|
    #  fd.integration_test(JSON.pretty_generate(JSON.load(@response.body)))
    #end

    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/organisations", headers: @headers
    end
    assert_response :ok
    integration_test("get-organisations.json")

    get "/v2/organisations/#{@org.id}", headers: @headers
    assert_response :ok
    integration_test("get-organisations-#{@org.id}.json")

    get "/v2/organisations/#{@org.id}?include[]=users", headers: @headers
    assert_response :ok
    integration_test("get-organisations-#{@org.id}-include[]=users.json")

    new_org = Organisation.new({
      name: "New Orga",
      invoice_name: "Someone",
      address: Address.new({
        street_and_number: "Teststr. 11",
        city: "Metropolis",
        postal_code: 12345,
        country: "Germany",
      }),
      vat_number: "DE345789003",
    })
    new_org.save!
    new_mem = Membership.new({
      user_id: @actor.user.id,
      organisation_id: new_org.id,
      role: :owner,
      status: :active,
      notify_device_update: true,
      notify_invoice: true
    })
    new_mem.save!
    svc_resp = OpenStruct.new({
      success: true,
      organisation: new_org,
      membership: new_mem,
    })

    OrganisationService.stub :create_organisation_and_subscription, svc_resp do
      post "/v2/organisations", headers: @headers, params: new_org.attributes
      assert_response :ok
      integration_test("post-organisations.json")
    end

    @org.memberships.each do |m|
      Stripe::Customer.stub :retrieve, "id" do
        get "/v2/memberships/#{m.id}", headers: @headers
      end
      assert_response :ok
      integration_test("get-memberships-#{m.id}.json")
    end

    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/memberships", headers: @headers
    end
    assert_response :ok
    integration_test("get-memberships.json")

    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/session", headers: @headers
      assert_response :ok
      integration_test("get-session.json")
      integration_test("post-session.json")
    end

    get "/v2/appconfig", headers: @headers
    assert_response :ok
    integration_test("get-appconfig.json")

    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/subscriptions/#{@org.subscription.id}", headers: @headers
    end
    assert_response :ok
    integration_test("get-subscriptions-#{@org.subscription.id}.json")

    Stripe::Customer.stub :retrieve, "id" do
      get "/v2/subscriptions/#{@org.subscription.id}/invoices", headers: @headers
    end
    assert_response :ok
    integration_test("get-subscriptions-#{@org.subscription.id}-invoices.json")

    @org.subscription.invoices.each do |i|
      get "/v2/subscriptions/#{@org.subscription.id}/invoices/#{i.id}", headers: @headers
      assert_response :ok
      integration_test("get-subscriptions-#{@org.subscription.id}-invoices-#{i.id}.json")
    end

    get "/v2/users", headers: @headers
    assert_response :ok
    integration_test("get-users.json")

    get "/v2/users?include=organisations", headers: @headers
    assert_response :ok
    integration_test("get-users-include-orgs.json")

    @org.users.each do |m|
      get "/v2/users/#{m.id}", headers: @headers, params: { "include" => "organisations" }
      assert_response :ok
      integration_test("get-users-#{m.id}.json")
    end
  end

  test "should handle quota update on address-less orgs" do
    attr = {
      "device_quota" => 100,
    }
    @org = organisations(:invalid_kai_org)
    @membership = memberships(:admin_at_admin_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |org, q|
      org.public_id == @org.public_id && q[:devices] == 100
    end

    EventService.stub :new, eventMock do
      patch "/v2/organisations/#{@org.id}", params: attr, headers: @headers
    end

    assert_response :ok
    assert_mock eventMock
  end

  test "should handle quota update" do
    attr = {
      "device_quota" => 100,
    }
    @membership = memberships(:admin_at_admin_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |org, q|
      org.public_id == organisations(:kais_org).public_id && q[:devices] == 100
    end

    EventService.stub :new, eventMock do
      patch "/v2/organisations/#{@org.id}", params: attr, headers: @headers
    end

    assert_response :ok
    assert_mock eventMock
    assert_equal JSON.load(@response.body).pretty_inspect, @response_body.pretty_inspect

    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |a,b| true end
    EventService.stub :new, eventMock do
      patch "/v2/organisations/#{@org.id}", params: @request_body.merge({:device_quota => "blah"}), headers: @headers
    end
    assert_response 422

    @membership = memberships(:kai_at_kais_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    EventService.stub :new, eventMock do
      patch "/v2/organisations/#{@org.id}", params: @request_body.merge({:device_quota => "blah"}), headers: @headers
    end
    assert_response :ok
  end
end
