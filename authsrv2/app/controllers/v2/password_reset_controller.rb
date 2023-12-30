module V2
  class PasswordResetController < V2::ApiBaseController
    skip_before_action :authenticate

    def create
      user = User.find_by_email(params[:email].downcase)
      # This line sends an email to the user with instructions on how to reset their password (a url with a random token)
      user.deliver_reset_password_instructions! if user
      # Tell the user instructions have been sent whether or not email was found.
      # This is to not leak information to attackers about which emails exist in the system.
      render json: { ok: true }, status: 200
    end

    def edit
      user = User.load_from_reset_password_token(params[:id])

      return render json: { error: :not_found }, status: 404 unless user.present?

      render json: { ok: true }, status: 200
    end

    def update
      user = User.load_from_reset_password_token(params[:id])

      return render json: { error: :not_found }, status: 404 unless user.present?

      begin
        # the next line makes the password confirmation validation work
        # user.password_confirmation = params[:password]
        # the next line clears the temporary token and updates the password
        user.change_password!(params[:password])
      rescue ActiveRecord::RecordInvalid
        render json: { errors: user.errors }, status: 422
      else
        render json: { ok: true }, status: 200
      end
    end
  end
end
