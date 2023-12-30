# Subjects
#
# Service: tag:immu.ne,2021:service/<service name>
# User: tag:immu.ne,2021:user/<uuid>
# Organisation: tag:immu.ne,2021:organisation/<uuid>
# Device: tag:immu.ne,2021:device/<root key name as hex>
# Agent tag:immu.ne,2021:agent
#
# Tokens
# {
#   "kid": first 8 chars of the hex SHA256 hash of the key (DER encoding)
# }
# {
#   "sub": org, user, service or device
#   "act": if org {
#     "sub": agent, user, service
#   },
#   "iss": service
#   "exp": future,
#   "nbf": now
#   "iat": now
# }

class TokenError < StandardError
end

class TokenService < ApplicationService
  ORGANISATION_TAG_PREFIX = "tag:immu.ne,2021:organisation/"
  SERVICE_TAG_PREFIX = "tag:immu.ne,2021:service/"
  USER_TAG_PREFIX = "tag:immu.ne,2021:user/"
  AGENT_TAG = "tag:immu.ne,2021:agent"

  @@private_key_path = Settings.authentication.private_key_path
  @@static_keyset = Settings.authentication.static_keys
  @@service_name = SERVICE_TAG_PREFIX + Settings.service_name
  @@api_token_lifetime = Settings.authentication.api_token_lifetime_seconds.to_i
  @@service_token_lifetime = Settings.authentication.service_token_lifetime_seconds.to_i
  @@enroll_token_lifetime = Settings.authentication.enrollment_token_lifetime_seconds.to_i
  @@key = nil
  @@kid = nil
  @@pub = nil

  def self.issue_enrollment_token(membership)
    load_key
    claims = self.default_claims(
      # Subject: the organisation
      subject: self.organisation_tag(membership.organisation_id),
      # Actor: agent
      actor: AGENT_TAG,
      # JWT ID: The memberships dynamic token
      jti: membership.jwt_token_key,
      # Expiration time
      lifetime: @@enroll_token_lifetime)
    header = {
      kid: @@kid,
    }

    JWT.encode claims, @@key, 'ES256', header
  end

  # Token used by the webapp to authenticate against this service and apisrv.
  # The token always represents a single user but can be scoped to a single
  # organisation (Membership argument).
  #
  # The apisrv cannot handle tokens with 'user' scope.
  def self.issue_api_token(membership)
    load_key
    # valid for the organisation only
    claims = self.default_claims(
      # Subject: the organisation
      subject: self.organisation_tag(membership.organisation_id),
      # Actor: agent
      actor: membership.user,
      # JWT ID: The memberships dynamic token
      jti: membership.jwt_token_key,
      # Expiration time
      lifetime: @@api_token_lifetime)

    header = {
      kid: @@kid,
    }

    JWT.encode claims, @@key, 'ES256', header
  end

  # Token used by this service to use the API server's private APIs
  def self.issue_service_token(organisation = nil)
    load_key
    case organisation
    when Organisation
      claims = self.default_claims(
        # Subject: the user
        subject: self.organisation_tag(organisation.public_id),
        actor: @@service_name,
        # JWT ID: The user dynamic token
        jti: SecureRandom.hex(32),
        # Expiration time
        lifetime: @@service_token_lifetime)

    when nil
      claims = self.default_claims(
        # Subject: the user
        subject: @@service_name,
        # JWT ID: The user dynamic token
        jti: SecureRandom.hex(32),
        # Expiration time
        lifetime: @@service_token_lifetime)

    else
      raise ArgumentError.new 'valid API token scopes are Organisation and nil'

    end
    header = {
      kid: @@kid,
    }

    JWT.encode claims, @@key, 'ES256', header
  end

  # Verifies an incoming Bearer token.
  # The function returns an Actor representing the token issuer or user
  def self.verify_token(token)
    load_key
    keyset = []
    claims = JWT.decode token, nil, false
    iss = claims && claims[0] && claims[0]['iss']
    raise TokenError.new 'illformed issuer' if !iss.delete_prefix! SERVICE_TAG_PREFIX

    self.get_keys(iss).each do |key|
      # verify token signature
      claims = JWT.decode token, key, true, { algorithm: 'ES256' }
      claims &&= claims[0]
      act = claims.fetch('act', {}).fetch('sub', '')
      sub = claims.fetch('sub', '')
      jti = claims.fetch('jti', '')

      raise TokenError.new 'Missing jti claim' if jti == ''

      # convert sub and act claims into Actor instance
      case sub
      when /^tag:immu\.ne,2021:organisation\/.+$/

        case act
        when /^tag:immu\.ne,2021:user\/.+$/
          actor = Actor.new(:user, membership: Membership.find_by!(jwt_token_key: jti))
          # sanity checks
          raise TokenError.new 'wrong subject' if self.organisation_tag(actor.organisation&.public_id) != sub
          raise TokenError.new 'wrong actor' if self.user_tag(actor.user&.public_id) != claims.fetch('act', {}).fetch('sub', '')
          return actor

        when /^tag:immu\.ne,2021:service\/.+$/
          act.delete_prefix! SERVICE_TAG_PREFIX
          raise TokenError.new 'wrong or missing issuer claim' if act != iss
          return Actor.new(
            :service, service: iss,
            organisation: Organisation.find_by!(public_id: sub.delete_prefix(ORGANISATION_TAG_PREFIX)))

        else
          raise TokenError.new 'unknown actor'

        end

      when /^tag:immu\.ne,2021:service\/.+$/
        raise TokenError.new 'wrong or missing issuer claim' if sub.delete_prefix(SERVICE_TAG_PREFIX) != iss
        return Actor.new(:service, service: iss)

      else
        raise TokenError.new 'invalid scope'
      end
    rescue JWT::DecodeError => e
      next
    end

    raise TokenError, 'No key verifies the token'
  rescue JWT::DecodeError
    raise TokenError.new 'Illformed token'
  end

  def self.load_key
    return if @@key != nil && @@kid != nil

    # auto generate an JWT signing key in dev
    if !@@private_key_path
      @@key = OpenSSL::PKey::EC.new 'prime256v1'
      @@key.generate_key
      @@pub = OpenSSL::PKey::EC.new @@key
      @@pub.private_key = nil
      @@kid = (Digest::SHA256.hexdigest @@pub.public_to_der)[0..15]
      b64_pubkey = @@pub.public_to_pem.lines[1..-2].map(&:strip).join
      puts "Service public key: #{b64_pubkey}"
    else
      File.open @@private_key_path do |fd|
        self.set_key fd.read
      end
    end
  end

  private

  def self.get_keys(issuer)
    # select the key set based on the issuer claim.
    #
    # XXX: make KeyDiscovery handle kids is we don't have to try all possible
    # public keys of an issuer.
    keyset = KeyDiscoveryService.keyset.fetch(issuer, [])
    if !Rails.env.production?
      keyset << @@pub
    end
    keyset + @@static_keyset.to_h.fetch(issuer.to_sym, []).map {|s| self.parse_key s}
  end

  def self.parse_key(b64)
    ec = OpenSSL::PKey::EC.new(Base64.decode64(b64))
    ec.check_key
    ec
  end

  def self.set_key(pkcs8)
    der = Base64.decode64 pkcs8
    @@key = OpenSSL::PKey.read(der)
    raise ArgumentError.new("Wrong curve") if @@key.group.curve_name != 'prime256v1'
    @@key.check_key
    @@pub = OpenSSL::PKey::EC.new 'prime256v1'
    @@pub.public_key = @@key.public_key
    @@kid = (Digest::SHA256.hexdigest @@pub.public_to_der)[0..15]
  end

  def self.user_tag(public_id)
    USER_TAG_PREFIX + public_id.to_s
  end

  def self.organisation_tag(public_id)
    ORGANISATION_TAG_PREFIX + public_id.to_s
  end

  def self.default_claims(subject:, actor: nil, jti:, lifetime:)
    case actor
    when String
      act = {
        act: {
          sub: actor
        }
      }
    when User
      act = {
        act: {
          sub: self.user_tag(actor.public_id),
          name: actor.name,
          rol: actor.role.to_s,
        }
      }
    when nil
      act = {}
    else
      raise TokenError.new
    end

    act.merge({
      sub: subject,
      jti: jti,
      # Expiration time
      exp: Time.now.to_i + lifetime,
      # Not before
      nbf: Time.now.to_i,
      # Issued at
      iat: Time.now.to_i,
      # Issuer, must match kid
      iss: @@service_name,
    })
  end
end
