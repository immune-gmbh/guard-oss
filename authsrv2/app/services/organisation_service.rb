class OrganisationService < ApplicationService

  def self.create_organisation_and_subscription(user, params)
    default_organisation_name = params[:name] || "#{user.name}'s Organisation"
    appendix = Organisation.where('name ilike ?', "%#{default_organisation_name}%").count
    default_organisation_name = "#{default_organisation_name}(#{appendix + 1})" if appendix > 0
    stripe_customer = nil
    stripe_subscription = nil

    ActiveRecord::Base.transaction do
      if !Settings.payment.disable
        stripe_customer = Stripe::Customer.create({
          email: user.email,
          name: default_organisation_name,
          description: params[:invoice_name],
          balance: user.memberships.role_owner.any? ? 0 : -Settings.payment.free_credits, # negative number indicates customer credit
          metadata: {
            created_by: user.id,
          }
        })
      end

      membership = Membership.new(
        user_id: user.id,
        role: :owner,
        status: :active,
        notify_device_update: true,
        notify_invoice: true
      )

      address = Address.create(
        street_and_number: params.dig(:address,:street_and_number),
        city: params.dig(:address,:city),
        postal_code: params.dig(:address,:postal_code),
        country: params.dig(:address,:country))

      organisation = Organisation.create(name: default_organisation_name) do |org|
        org.status = 'active'
        org.stripe_customer_id = stripe_customer&.id
        org.invoice_name = params[:invoice_name]
        org.vat_number = params[:vat_number]
        org.memberships << membership
        org.address = address if address&.persisted?
      end

      if !Settings.payment.disable
        stripe_subscription = Stripe::Subscription.create({
          customer: organisation.stripe_customer_id,
          items: [
            {
              price: Settings.payment.price_id # Stripe price id, full devices per month
            },
          ],
          metadata: {
            subscribed_by_membership: membership.id
          }
        })

        subscription = Subscription.create(
          organisation: organisation,
          stripe_subscription_id: stripe_subscription.id,
          stripe_subscription_item_id: stripe_subscription.items.data[0].id,
          period_start: Time.at(stripe_subscription.current_period_start).to_date,
          period_end: Time.at(stripe_subscription.current_period_end).to_date,
          new_devices_amount: 0
        )
      end

      # update org quota on other services
      EventService.new.update_quota organisation, devices: subscription&.max_devices_amount || Settings.payment.device_quota

      current_ability = Ability.new(Actor.new(:user, user: user))

      response({ success: true, organisation: organisation, membership: membership })
    end # ActiveRecord::Base.transaction
  rescue => e
    if !Settings.payment.disable
      Stripe::Subscription.delete(stripe_subscription.id) if stripe_subscription
      Stripe::Customer.delete(stripe_customer.id) if stripe_customer
    end

    response({ success: false, message: e&.message || e.inspect})
  end

  def self.update_address_and_taxes_with_stripe(organisation)
    return response({ success: true }) if Settings.payment.disable
    address = organisation.address

    Stripe::Customer.update(organisation.stripe_customer_id, {
      name: organisation.name,
      description: organisation.invoice_name,
      address: {
        city: address.city,
        country: ISO3166::Country.find_country_by_name(address.country).alpha2,
        line1: address.street_and_number,
        postal_code: address.postal_code,
      }
    })

    if address.country == "Germany" || (VatChecker.apply_vat?(address.country) && !organisation.vat_number.present?)
      tax_rate = Stripe::TaxRate
        .list({active: true, limit: 100})
        .detect {|x| x.country == ISO3166::Country.find_country_by_name(address.country).alpha2}

      stripe_subscription = Stripe::Subscription.update(organisation.subscription.stripe_subscription_id, {
        default_tax_rates: [ tax_rate ],
      })

    else
      # check if taxes are applied already and remove them (e.g. company relocated to non-EU countries)
      stripe_subscription = Stripe::Subscription.retrieve(organisation.subscription.stripe_subscription_id)

      if stripe_subscription.default_tax_rates.present?
        stripe_subscription = Stripe::Subscription.update(organisation.subscription.stripe_subscription_id, {
          default_tax_rates: nil,
        })
      end
    end

    response({ success: true })
  rescue => e
    response({ success: false, message: e&.message || e.inspect })
  end

  def self.update_vat_number_with_stripe(organisation)
    return response({ success: true }) if Settings.payment.disable

    # remove existing vat ids, if any
    tax_ids = Stripe::Customer.list_tax_ids(organisation.stripe_customer_id, {limit:3})
    tax_ids.map do |tax_id|
      Stripe::Customer.delete_tax_id(
        organisation.stripe_customer_id,
        tax_id.id
      )
    end

    stripe_vat_id = Stripe::Customer.create_tax_id(
      organisation.stripe_customer_id,
      {
        #type: VatChecker.stripe_taxid_for_country(organisation.address&.country, organisation.vat_number),
        type: 'eu_vat',
        value: organisation.vat_number
      },
    )
    response({ success: true, stripe_vat_id: stripe_vat_id })
  rescue => e
    response({ success: false, message: e&.message || e.inspect })
  end

  def self.create_organisation_for_invite(params)
    ActiveRecord::Base.transaction do
      default_organisation_name = params[:name]
      appendix = Organisation.where('name ilike ?', "%#{default_organisation_name}%").count
      default_organisation_name = "#{default_organisation_name}(#{appendix + 1})" if appendix > 0

      address = Address.create(
        street_and_number: params.dig(:address,:street_and_number),
        city: params.dig(:address,:city),
        postal_code: params.dig(:address,:postal_code),
        country: params.dig(:address,:country))
      organisation = Organisation.create(
        name: default_organisation_name,
        address: address.persisted? ? address : nil,
        freeloader: params[:freeloader] || false)

      # tell everyone else about the new org
      EventService.new.update_quota organisation, devices: 0

      response({ success: true, organisation: organisation})
    rescue => e
      response({ success: false, message: e})
    end
  end

  def self.invite_user_to_organisation(email, organisation_id, role, name = nil)
    ActiveRecord::Base.transaction do
      success = false
      message = 'Organisation not found or wrong id.'
      organisation = Organisation.find_by(public_id: organisation_id)

      if !organisation.present?
        response({ success: success, message: message })
      else
        user = User.find_or_create_by(email: email.downcase) do |user|
          user.name = name || 'Anonymous'
          user.invited = true
        end

        if user.valid?
          begin
            membership = Membership.create!(
              organisation: organisation,
              user: user,
              role: role || 'user'
            )

            UserMailer.invitation_email(membership).deliver_now
            message = 'Invitation has been sent.'
            success = true
          rescue ActiveRecord::RecordNotUnique => invalid
            message = 'User is already a member.'
            membership = Membership.find_by(organisation: organisation, user: user)
          end
          serialized_membership = MembershipSerializer.new(membership)
          serialized_user = UserSerializer.new(user)

          response({ success: success, message: message, user: serialized_user, membership: serialized_membership })
        else
          message = "User is invalid."
          serialized_user = UserSerializer.new(user)
          response({ success: success, message: message, user: serialized_user })
        end
      end
    end
  rescue => e
    response({ success: false, message: e&.message || e.inspect})
  end

  def self.create_subscription(owner_membership)
    owner = owner_membership.user
    organisation = owner_membership.organisation

    return response({ success: false, message: 'customer exists already' }) if organisation.stripe_customer_id
    return response({ success: false, message: 'subscription exists already' }) if organisation.subscription
    return response({ success: false, message: 'payments disabled' }) if Settings.payment.disable

    stripe_customer = nil
    stripe_subscription = nil

    stripe_customer = Stripe::Customer.create({
      email: owner.email,
      name: organisation.name,
      balance: owner.memberships.role_owner.count != 1 ? 0 : -Settings.payment.free_credits, # negative number indicates customer credit
      metadata: {
        created_by: owner.id,
      }
    })

    organisation.update(stripe_customer_id: stripe_customer.id)

    stripe_subscription = Stripe::Subscription.create({
      customer: organisation.stripe_customer_id,
      items: [
        {
          price: Settings.payment.price_id # Stripe price id, full devices per month
        },
      ],
      metadata: {
        subscribed_by_membership: owner_membership.id
      }
    })

    ActiveRecord::Base.transaction do
      subscription = Subscription.create(
        organisation: organisation,
        stripe_subscription_id: stripe_subscription.id,
        stripe_subscription_item_id: stripe_subscription.items.data[0].id,
        period_start: Time.at(stripe_subscription.current_period_start).to_date,
        period_end: Time.at(stripe_subscription.current_period_end).to_date,
        new_devices_amount: 0
      )

      # update org quota on other services
      EventService.new.update_quota organisation, devices: subscription.max_devices_amount

      response({ success: true, subscription: subscription })
    end
  rescue => e
    Stripe::Subscription.delete(stripe_subscription.id) if stripe_subscription
    Stripe::Customer.delete(stripe_customer.id) if stripe_customer

    response({ success: false, message: e&.message || e.inspect})
  end

  def self.setup_payment(organisation)
    raise RuntimeError.new 'organisation has no customer' if !organisation.stripe_customer_id
    raise RuntimeError.new 'organisation has no subscription' if !organisation.subscription
    raise RuntimeError.new 'payments disabled' if Settings.payment.disable

    setup_intent = Stripe::SetupIntent.create({
      customer: organisation.stripe_customer_id,
    })

    response({ success: true, setup_intent: setup_intent, subscription: organisation.subscription })
  rescue => e
    response({ success: false, message: e&.message || e.inspect })
  end

  def self.make_payment_method_default(organisation, payment_method_id)
    raise RuntimeError.new 'organisation has no customer' if !organisation.stripe_customer_id
    raise RuntimeError.new 'organisation has no subscription' if !organisation.subscription

    if !Settings.payment.disable
      stripe_response = Stripe::Customer.update(organisation.stripe_customer_id, {
        invoice_settings: {
          default_payment_method: payment_method_id,
        }
      })
    end
    response({ success: true , subscription: organisation.subscription})
  rescue => e
    response({ success: false, message: e&.message || e.inspect })
  end

  def self.delete(organisation:, soft_delete:)
    ActiveRecord::Base.transaction do
      users = organisation.memberships.map(&:user)
      subscription = organisation.subscription
      stripe_subscription_id =
        subscription.stripe_subscription_id if subscription.present? && !organisation.deleted?

      # "delete" the organisation everywhere else
      EventService.new.update_quota organisation, devices: 0

      if !soft_delete
        organisation.destroy
      else
        organisation.deleted!
        organisation.memberships.destroy_all
        subscription.update(status: 'canceled') if subscription&.status == 'active'
      end

      users.map do |user|
        user.destroy if user.memberships.empty?
      end

      # no reasonable way to undo that, so do that last. worst case the tx
      # fails and we retain the db entries.
      if stripe_subscription_id && !Settings.payment.disable
        stripe_subscription = Stripe::Subscription.retrieve(stripe_subscription_id)
        if stripe_subscription.status != 'canceled'
          Stripe::Subscription.delete(stripe_subscription_id, { invoice_now: true })
        end
      end
    end
  end

  private

  def self.response(params)
    ro = Struct.new(*params.keys).new(*params.values)
  end
end
