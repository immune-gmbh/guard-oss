require "minitest/autorun"
require "test_helper"

class TokensTest < ActiveSupport::TestCase
  test "we can verify our user tokens" do
    token = TokenService.issue_api_token(memberships(:kai_at_kais_org))

    act = TokenService.verify_token(token)
    assert act.user == users(:kai)
    assert act.organisation == organisations(:kais_org)

    abl = Ability.new act
    assert abl.can? :manage, users(:kai)
    assert abl.can? :manage, organisations(:kais_org)
    assert abl.can? :read, organisations(:kais_org)
    assert abl.can? :read, organisations(:testys_org)
    assert abl.cannot? :manage, organisations(:testys_org)
  end

  test "we can verify our membership tokens" do
    token = TokenService.issue_api_token(users(:kai).memberships.first)

    act = TokenService.verify_token(token)
    assert act.user == users(:kai)
    assert act.organisation == users(:kai).memberships.first.organisation

    abl = Ability.new act
    assert abl.can :manage, users(:kai)
    assert abl.can(:read, organisations(:kais_org))
    assert abl.can(:manage, organisations(:kais_org))
    assert abl.can(:read, organisations(:testys_org))
    assert abl.cannot? :manage, organisations(:testys_org)
  end

  test "we handle illformed tokens" do
    assert_raise do
      TokenService.verify_token("lallalalala")
    end
  end

  test "missing jti" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      sub: "tag:immu.ne,2021:organisation/#{organisations(:kais_org).public_id}",
      act: {
        sub: 'tag:immu.ne,2021:service/authsrv',
      },
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "illformed issuer" do
    TokenService.load_key

    claims = {
      iss: 'authsrv',
      jti: SecureRandom.hex(32),
      sub: "tag:immu.ne,2021:organisation/#{organisations(:kais_org).public_id}",
      act: {
        sub: 'tag:immu.ne,2021:service/authsrv',
      },
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "unknown actor" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      jti: SecureRandom.hex(32),
      sub: "tag:immu.ne,2021:organisation/#{organisations(:kais_org).public_id}",
      act: {
        sub: 'authsrv',
      },
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "illformed actor" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      jti: SecureRandom.hex(32),
      sub: "tag:immu.ne,2021:organisation/#{organisations(:kais_org).public_id}",
      act: {
        sub: 'tag:immu.ne,2021:service/apisrv',
      },
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "wrong actor/subject pair" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      jti: SecureRandom.hex(32),
      sub: "tag:immu.ne,2021:service/apisrv",
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "unknown subject" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      jti: SecureRandom.hex(32),
      sub: "blah",
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {
      kid: TokenService.class_variable_get(:@@kid),
    }

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenService.verify_token token
    end
  end

  test "unknown kid" do
    TokenService.load_key

    claims = {
      iss: 'tag:immu.ne,2021:service/authsrv',
      jti: SecureRandom.hex(32),
      sub: 'tag:immu.ne,2021:service/authsrv',
      exp: (Time.now + 24.hour).to_i,
      nbf: Time.now.to_i,
    }
    header = {}

    assert_raise do
      token = JWT.encode(claims, TokenService.class_variable_get(:@@key), 'ES256', header)
      TokenSerivce.class_variable_set(:@@key, nil)
      TokenService.load_key
      TokenService.verify_token token
    end
  end

  test "broken signature" do
    token = TokenService.issue_service_token(organisations(:kais_org))
    token[42] = (token[42].ord ^ 0xff).chr

    assert_raise do
      TokenService.verify_token token
    end
  end

  test "broken header" do
    token = TokenService.issue_service_token(organisations(:kais_org))
    token[0] = (token[0].ord ^ 0xff).chr

    assert_raise do
      TokenService.verify_token token
    end
  end

  test "we can verify our service tokens (org)" do
    token = TokenService.issue_service_token(organisations(:kais_org))

    act = TokenService.verify_token(token)
    assert act.organisation == organisations(:kais_org)
    assert act.user == nil
    assert act.service == "authsrv"

    abl = Ability.new act
    assert abl.can(:alert, organisations(:kais_org))
    assert abl.cannot? :manage, organisations(:kais_org)
  end

  test "we can verify our service tokens (wildcard)" do
    token = TokenService.issue_service_token(nil)

    act = TokenService.verify_token(token)
    assert act.organisation == nil
    assert act.user == nil
    assert act.service == "authsrv"

    abl = Ability.new act
    assert abl.can(:alert, organisations(:kais_org))
    assert abl.cannot? :manage, organisations(:kais_org)
  end

  test "we can verify apisrv service tokens" do
    pkcs8 = file_fixture("token.key").read
    der = Base64.decode64 pkcs8
    key = OpenSSL::PKey.read(der)
    raise ArgumentError.new("Wrong curve") if key.group.curve_name != 'prime256v1'
    key.check_key
    pub = OpenSSL::PKey::EC.new 'prime256v1'
    pub.public_key = key.public_key
    keyset = {"apisrv" => [pub]}

    KeyDiscoveryService.keyset = keyset
    base = Time.parse "2021-08-31 17:14:05.24705 +0000 UTC"
    tokens = [
      file_fixture("token1").read,
      file_fixture("token2").read,
      file_fixture("token3").read,
    ]

    Time.stub :now, base + 10.seconds do
      tokens.each do |tok|
        act = TokenService.verify_token tok

        assert act.user == nil
        assert act.service == "apisrv"
        assert act.organisation == organisations(:fixed_id_org)
      end
    end
  end

  test "issue service token to user" do
    assert_raise ArgumentError do
      TokenService.issue_service_token users(:kai)
    end
  end

  test "gen apisrv fixtures" do
    TokenService.class_variable_set(:@@api_token_lifetime, 60 * 60 * 24 * 356 * 10)
    TokenService.class_variable_set(:@@service_token_lifetime, 60 * 60 * 24 * 356 * 10)
    namespaced_svc_tok = TokenService.issue_service_token(organisations(:kais_org))
    write_fixture("namespaced-service-token", namespaced_svc_tok)

    usr_tok = TokenService.issue_api_token(memberships(:kai_at_kais_org))
    write_fixture("user-token", usr_tok)

    svc_tok = TokenService.issue_service_token(nil)
    write_fixture("service-token", svc_tok)

    pub = TokenService.class_variable_get(:@@pub)
    write_fixture("token-public-key", pub.public_to_pem.lines[1..-2].map(&:strip).join)
  end
end
