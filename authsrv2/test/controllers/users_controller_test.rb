require "minitest/autorun"
require "test_helper"

class UsersControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = users(:kai)
    @membership = memberships(:kai_at_kais_org)
    @actor = Actor.new(:user, membership: @membership)
    @ability = Ability.new @actor
    @request_body = @user.serializable_hash
    @request_body["address"] = @user.address.serializable_hash
    @response_body = UserSerializer.new(@user, { include: [:address], params: {current_ability: @ability }}).as_json
    @headers = {
      "Authorization" => "Bearer #{TokenService.issue_api_token @membership}",
      "Accept" => "application/vnd.api+json"
    }
    @svc_ok = OpenStruct.new({ success: true })
    @svc_err = OpenStruct.new({ success: false, message: "blah" })
  end

  test "get single user" do
    get "/v2/users/#{@user.id}", headers: @headers
    assert_response :ok
    assert_equal @response_body.pretty_inspect, JSON.load(@response.body).pretty_inspect

    get "/v2/users/xxx", headers: @headers
    assert_response :not_found

    get "/v2/users/#{@user.id}"
    assert_response :unauthorized

    get "/v2/users/#{users(:testy).id}", headers: @headers
    assert_response :forbidden
  end

  test "list users" do
    get "/v2/users", headers: @headers
    assert_response :ok
    assert_equal @membership.organisation.users.size, JSON.load(@response.body)["data"].size
    integration_test("get-users.json")
  end

  test "patch single user" do
    @request_body["has_seen_intro"] = true
    patch "/v2/users/#{@user.id}", headers: @headers, params: @request_body
    assert_response :ok
    @response_body["data"]["attributes"]["has_seen_intro"] = true
    assert_equal @response_body.pretty_inspect, JSON.load(@response.body).pretty_inspect
  end

  test "change my role" do
    @membership = memberships(:admin_at_admin_org)
    @headers["Authorization"] = "Bearer #{TokenService.issue_api_token @membership}"

    @request_body["role"] = "admin"
    patch "/v2/users/#{@user.id}", headers: @headers, params: @request_body
    assert_response :ok
    @response_body["data"]["attributes"]["role"] = "admin"
    assert_equal @response_body.pretty_inspect, JSON.load(@response.body).pretty_inspect
  end

  test "can't make myself admin" do
    @request_body["role"] = "admin"
    patch "/v2/users/#{@user.id}", headers: @headers, params: @request_body
    assert_response :ok
    assert_equal @response_body.pretty_inspect, JSON.load(@response.body).pretty_inspect
  end

  test "failed patch" do
    @request_body["name"] = ""
    patch "/v2/users/#{@user.id}", headers: @headers, params: @request_body
    assert_response :unprocessable_entity
  end

  test "register new user -- payment disabled, activation turned off" do 
    @request_body = {
      "name"  =>"Test User",
      "email" => "testtest@example.com",
      "password" => "blah",
    }

    Settings.authentication.disable_activation = true
    Settings.payment.disable = true
    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |org, q| q[:devices] > 0 end

    EventService.stub :new, eventMock do
      assert_emails 0 do
        post "/v2/users", params: @request_body
        assert_response :ok
      end
    end
    Settings.authentication.disable_activation = false
    Settings.payment.disable = false

    resp = JSON.load(@response.body)
    assert_equal resp["data"]["attributes"]["name"], "Test User"
    assert_equal resp["data"]["attributes"]["email"], "testtest@example.com"
    
    new_user = User.find(resp["data"]["id"])
    assert_equal new_user.activation_state, "active"
    assert_equal 1, new_user.organisations.size
    assert_equal 1, new_user.memberships.size
  end
 
  test "register new user with activation turned off" do 
    @request_body = {
      "name"  =>"Test User",
      "email" => "testtest@example.com",
      "password" => "blah",
    }

    Settings.authentication.disable_activation = true
    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |org, q| q[:devices] > 0 end
 
    sub = OpenStruct.new({
      id: "id",
      items: OpenStruct.new({
        data: [OpenStruct.new({
          id: "id"
        })],
      }),
      current_period_start: Time.now(),
      current_period_end: Time.now(),
    })
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, eventMock do
        Stripe::Customer.stub :create, OpenStruct.new({id: "id"}) do 
          Stripe::Subscription.stub :create, sub do 
            assert_emails 0 do
              post "/v2/users", params: @request_body
              assert_response :ok
            end
          end
        end
      end
    end
    Settings.authentication.disable_activation = false

    resp = JSON.load(@response.body)
    assert_equal resp["data"]["attributes"]["name"], "Test User"
    assert_equal resp["data"]["attributes"]["email"], "testtest@example.com"
    
    new_user = User.find(resp["data"]["id"])
    assert_equal new_user.activation_state, "active"
    assert_equal 1, new_user.organisations.size
    assert_equal 1, new_user.memberships.size
  end
 
  test "register new user" do 
    @request_body = {
      "name"  =>"Test User",
      "email" => "testtest@example.com",
      "password" => "blah",
    }

    assert_emails 1 do
      post "/v2/users", params: @request_body
      assert_response :ok
    end

    resp = JSON.load(@response.body)
    assert_equal resp["data"]["attributes"]["name"], "Test User"
    assert_equal resp["data"]["attributes"]["email"], "testtest@example.com"
    
    new_user = User.find(resp["data"]["id"])
    assert_equal new_user.activation_state, "pending"

    assert_emails 1 do
      post "/v2/users/#{new_user.id}/resend", params: { email: "testtest@example.com" }
      assert_response :ok
    end

    OrganisationService.stub :create_organisation_and_subscription, @svc_err do 
      post "/v2/users/#{new_user.activation_token}/activate"
      assert_response :service_unavailable
    end
    eventMock = Minitest::Mock.new
    eventMock.expect(:update_quota, true) do |org, q| q[:devices] > 0 end
 
    sub = OpenStruct.new({
      id: "id",
      items: OpenStruct.new({
        data: [OpenStruct.new({
          id: "id"
        })],
      }),
      current_period_start: Time.now(),
      current_period_end: Time.now(),
    })
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, eventMock do
        Stripe::Customer.stub :create, OpenStruct.new({id: "id"}) do 
          Stripe::Subscription.stub :create, sub do 
            post "/v2/users/#{new_user.activation_token}/activate"
            assert_response :ok
          end
        end
      end
    end

    assert_emails 0 do
      post "/v2/users/#{new_user.id}/resend", params: { email: "testtest@example.com" }
      assert_response :ok
    end

    new_user.reload
    new_user.organisations.reload
    new_user.memberships.reload
    assert_equal new_user.activation_state, "active"
    assert_equal 1, new_user.organisations.size
    assert_equal 1, new_user.memberships.size
  end

  test "fail to active new user" do 
    OrganisationService.stub :create_organisation_and_subscription, @svc_ok do 
      post "/v2/users/blah/activate"
      assert_response :not_found
    end
  end

  test "fail to register new user" do 
    @request_body = {
      "name"  =>"Test User",
      "email" => "test@example.com",
      "password" => "blah",
    }
    post "/v2/users", params: @request_body
    assert_response :unprocessable_entity

    @request_body["email"] = "testtest@example.com"
    @request_body["name"] = ""
    post "/v2/users", params: @request_body
    assert_response :unprocessable_entity

    @request_body["password"] = ""
    @request_body["name"] = "LAlalalalak"
    post "/v2/users", params: @request_body
    assert_response :unprocessable_entity
  end

  test "delete user" do
    skip "todo"   
  end
end
