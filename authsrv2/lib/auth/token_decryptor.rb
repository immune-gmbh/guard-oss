module Auth::TokenDecryptor
  extend self

  def call(token)
    decrypt(token)
  end

  private

  def decrypt(token)
    JWT.decode(token, Rails.application.credentials.secret_key_base)
  rescue JWT::DecodeError
    raise InvalidTokenError
  end
end

class InvalidTokenError < StandardError; end
