require 'minitest/autorun'
require "test_helper"

class SessionsControllerTest < ActionDispatch::IntegrationTest
  fixtures :users

  setup do
    @eventMock = Minitest::Mock.new
    @eventMock.expect(:revoke_token, true) do |mem| true end
  end
 
  test "should login via oauth2 token" do
    def write(str)
      File.open("test/fixtures/output/#{str}", "w+") do |fd|
        fd.write(JSON.pretty_generate(JSON.load(@response.body)))
      end
    end

    users(:kai).prepare_token_login
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        assert_changes "users(:kai).login_token" do
          post '/v2/session', params: { token: users(:kai).login_token }
          write("post-session.json")
          assert_response :success
          users(:kai).reload
        end
      end
    end
  end

  test "session includes memberships and tokens" do
    skip
  end

  test "should fail login with wrong token" do
    EventService.stub :new, @eventMock do
      post '/v2/session', params: { token: "test-token" }
    end
    assert_response :unauthorized
  end

  test "should logout" do
    assert_changes("users(:kai).memberships.map(&:jwt_token_key)") do
      EventService.stub :new, @eventMock do
        delete '/v2/session', headers: { Authorization: "Bearer #{TokenService.issue_api_token(users(:kai).memberships.first)}" }
      end
      assert_response :ok
      users(:kai).memberships.first.reload
    end
  end

  test "should fail logout w/o token" do
    EventService.stub :new, @eventMock do
      delete '/v2/session'
    end
    assert_response :ok
  end

  test "should login with correct email/pw" do
    user = users(:kai)
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        post '/v2/session', params: { email: user.email, password: 'secret' }
      end
    end
    assert_response :ok

    tokens = JSON.load(body)['included'].select do |x| x["type"] == "membership" end
    assert_not_empty tokens
  end

  test "should not login with bad email/pw" do
    user = users(:kai)
    EventService.stub :new, @eventMock do
      post '/v2/session', params: { email: user.email, password: 'wrong' }
    end
    assert_response :unauthorized

    EventService.stub :new, @eventMock do
      post '/v2/session', params: { email: 'non-existent@example.com', password: 'wrong' }
    end
    assert_response :unauthorized
  end

  test "should not login if not activated" do
    user = users(:not_active)
    EventService.stub :new, @eventMock do
      post '/v2/session', params: { email: user.email, password: 'secret' }
    end
    assert_response :forbidden
  end

  test "should not login if not address was given" do
    user = users(:no_address)
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        post '/v2/session', params: { email: user.email, password: 'secret' }
      end
    end
    assert_response :ok

    assert_equal JSON.load(body)['data']['attributes']['next_path'], "https://xxxx.xxxxx/registration/activate_email?email=xxxx.xxxx@xxxx.xxxx"
  end

  test "should handle user w/o org" do
    user = users(:no_org)
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        post '/v2/session', params: { email: user.email, password: 'secret' }
      end
    end
    assert_response :ok
    assert_empty (JSON.load(body).dig("included").select do |x| x["type"] == "membership" end).to_a
  end

  test "should refresh session" do
    user = users(:kai)
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        post '/v2/session', params: { email: user.email, password: 'secret' }
      end
    end
    assert_response :ok
    tokens = JSON.load(body)["included"].select do |x| x["type"] == "membership" end

    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        get '/v2/session/refresh', headers: { "Authorization" => "Bearer #{tokens.dig(0, "attributes", "token")}" }
      end
    end
    assert_response :ok
  end

  test "shouldn't refresh bad session" do
    EventService.stub :new, @eventMock do
      get '/v2/session/refresh'
    end
    assert_response :unauthorized

    user = users(:kai)
    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        post '/v2/session', params: { email: user.email, password: 'secret' }
      end
    end
    assert_response :ok

    Stripe::Customer.stub :retrieve, "id" do
      EventService.stub :new, @eventMock do
        delete '/v2/session'
      end
    end
    assert_response :ok

    EventService.stub :new, @eventMock do
      get '/v2/session/refresh'
    end
    assert_response :unauthorized
  end
end
