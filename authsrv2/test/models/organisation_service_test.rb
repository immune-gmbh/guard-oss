require "minitest/autorun"
require "test_helper"
require 'stripe_mock'

module StripeMock
  module RequestHandlers
    module Subscriptions
      def verify_card_present(customer, plan, subscription, params={})
        nil
      end
    end
  end
end


class OrganisationServiceTest < ActiveSupport::TestCase
  setup do
    StripeMock.start
    @product = Stripe::Product.create({name: 'Gold Special'})
    @price = Stripe::Price.create({
      unit_amount: 10,
      currency: 'usd',
      recurring: {interval: 'month'},
      product: @product.id,
    })
    Settings.payment.price_id = @price.id
    @eventMock = Minitest::Mock.new
    @eventMock.expect(:update_quota, true) do |mem| true end

    JSON.load(file_fixture('tax-rates.json').read).each do |t| 
      Stripe::TaxRate.create(t)
    end
  end

  teardown do
    StripeMock.stop
  end

  test "create orga for new user" do
    user = User.create({
      name: "Test2",
      email: "test2@example.com",
      has_seen_intro: false,
      password: "123",
    })
    user.activate!

    EventService.stub :new, @eventMock do
      response = OrganisationService.create_organisation_and_subscription(user, {})
      assert response&.success
    end

    user.reload
    assert user.organisations.size == 1
    assert user.organisations.first&.subscription
    assert_nil user.organisations.first.address

    customer = Stripe::Customer.retrieve(user.organisations.first.stripe_customer_id)
    assert customer.balance == -Settings.payment.free_credits
  end

  test "2nd orga has no free credits" do
    kai = users(:kai)
    EventService.stub :new, @eventMock do
      response = OrganisationService.create_organisation_and_subscription(kai, {})
      assert response&.success
    end

    kai.reload
    customer = Stripe::Customer.retrieve(kai.organisations.last.stripe_customer_id)
    assert customer.balance == 0
  end

  test "create orga with address" do
    kai = users(:kai)
    address = {
      street_and_number: "Teststr. 7",
      city: "Metropolis",
      postal_code: "123456",
      country: "DE",
    }

    EventService.stub :new, @eventMock do
      response = OrganisationService.create_organisation_and_subscription(kai, {address: address})
      assert response&.success
    end

    kai.reload
    assert_operator kai.organisations.last.address.attributes.symbolize_keys, :>=, address
  end

  test "create orga handles stripe error" do
    kai = users(:kai)

    StripeMock.prepare_error(StandardError.new("Test error"), :new_customer)
    response = OrganisationService.create_organisation_and_subscription(kai, {})
    assert !response&.success
    assert_includes response.message, "Test error"

    StripeMock.prepare_error(StandardError.new("Test error"), :create_subscription)
    response = OrganisationService.create_organisation_and_subscription(kai, {})
    assert !response&.success
    assert_includes response.message, "Test error"
  end

  test "handle address update" do
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id
    response = OrganisationService.update_address_and_taxes_with_stripe(org)
    assert response&.success

    address = Stripe::Customer.retrieve(org.stripe_customer_id).address
    assert_equal org.address.street_and_number, address.line1
    assert_equal org.address.city, address.city
    assert_equal org.address.postal_code, address.postal_code
    assert_equal ISO3166::Country.find_country_by_name(org.address.country).alpha2, address.country
  end

  test "vat german customer" do
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id
    response = OrganisationService.update_address_and_taxes_with_stripe(org)
    assert response&.success

    subscription = Stripe::Subscription.retrieve(org.subscription.stripe_subscription_id)
    assert_equal subscription.default_tax_rates.size, 1
    assert_equal subscription.default_tax_rates.first&.jurisdiction, "DE"
  end

  test "vat EU customer" do
    org = organisations(:kais_org)
    org.address.country = "Finland"
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id
    response = OrganisationService.update_address_and_taxes_with_stripe(org)
    assert response&.success

    subscription = Stripe::Subscription.retrieve(org.subscription.stripe_subscription_id)
    assert_equal subscription.default_tax_rates.size, 1
    assert_equal subscription.default_tax_rates.first&.jurisdiction, "FI"
  end

  test "vat non-EU customer" do
    org = organisations(:kais_org)
    org.address.country = "United States"
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id
    response = OrganisationService.update_address_and_taxes_with_stripe(org)
    assert response&.success

    subscription = Stripe::Subscription.retrieve(org.subscription.stripe_subscription_id)
    assert_nil subscription.default_tax_rates
  end

  test "vat company w/ VAT ID" do
    org = organisations(:kais_org)
    org.vat_number = "123456"
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id
    response = OrganisationService.update_address_and_taxes_with_stripe(org)
    assert response&.success

    subscription = Stripe::Subscription.retrieve(org.subscription.stripe_subscription_id)
    assert_equal subscription.default_tax_rates.first&.jurisdiction, "DE"
  end

  test "update VAT ID" do
    # https://github.com/stripe-ruby-mock/stripe-ruby-mock/issues/660
    skip
  end

  test "create invite only orga" do
    EventService.stub :new, @eventMock do
      response = OrganisationService.create_organisation_for_invite({
        name: "Test invite orga"
      })
      assert response&.success
    end
  end

  test "create invite only orga w/ address" do
    address = {
      street_and_number: "Teststr. 7",
      city: "Metropolis",
      postal_code: "123456",
      country: "DE",
    }
    EventService.stub :new, @eventMock do
      response = OrganisationService.create_organisation_for_invite({
        name: "Test invite orga",
        address: address,
      })
      assert response&.success
      assert_operator response.organisation.address.attributes.symbolize_keys, :>=, address
    end
  end

  test "invite to existing org" do
    org = organisations(:kais_org) 
    resp = OrganisationService.invite_user_to_organisation("lalala@example.com", org.id, "user", "New invited user")
    assert resp.success
    user = User.find_by(email:"lalala@example.com")
    assert user.invited
  end

  test "invite existing user" do
    org = organisations(:kais_org) 
    resp = OrganisationService.invite_user_to_organisation("kai.michaelis@immu.ne", org.id, "user", "New invited user")
    assert !resp.success
  end

  test "invite to non-existent org" do
    resp = OrganisationService.invite_user_to_organisation("test1234@example.com", "blah", "user", "New invited user")
    assert !resp.success
  end

  test "invite with invalid role" do
    org = organisations(:kais_org) 
    resp = OrganisationService.invite_user_to_organisation("test1234@example.com",org.id, "blah", "New invited user")
    assert !resp.success
  end

  test "invite invalid user" do
    org = organisations(:kais_org) 
    resp = OrganisationService.invite_user_to_organisation("test1234@example.com",org.id, nil, " ")
    assert !resp.success
  end

  test "create subscription on new org" do  
    EventService.stub :new, @eventMock do
      OrganisationService.create_organisation_for_invite({
        name: "Test invite orga"
      })
    end
    org = Organisation.find_by(name: "Test invite orga")
    OrganisationService.invite_user_to_organisation("test1234@example.com",org.id, nil, "New invited user")

    mem = User.find_by(name: "New invited user").memberships.first
    resp = OrganisationService.create_subscription(mem)
    assert resp.success
  end

  test "create subscription on active org" do 
    mem = memberships(:kai_at_kais_org)
    resp = OrganisationService.create_subscription(mem)
    assert !resp.success

    mem.organisation.stripe_customer_id = nil
    resp = OrganisationService.create_subscription(mem)
    assert !resp.success
 end

  test "handle stripe error on create subscription" do
    EventService.stub :new, @eventMock do
      OrganisationService.create_organisation_for_invite({
        name: "Test invite orga"
      })
    end
    org = Organisation.find_by(name: "Test invite orga")
    OrganisationService.invite_user_to_organisation("test1234@example.com",org.id, nil, "New invited user")

    mem = User.find_by(name: "New invited user").memberships.first
    StripeMock.prepare_error(StandardError.new("Test error"), :new_customer)
    resp = OrganisationService.create_subscription(mem)
    assert !resp.success
    assert_includes resp.message, "Test error"

    StripeMock.prepare_error(StandardError.new("Test error"), :create_subscription)
    resp = OrganisationService.create_subscription(mem)
    assert !resp.success
    assert_includes resp.message, "Test error"
  end

  test "setup payment" do 
    resp = OrganisationService.setup_payment(organisations(:kais_org))
    assert resp.success
   end

  test "setup payment for org w/o customer" do 
    EventService.stub :new, @eventMock do
    OrganisationService.create_organisation_for_invite({
        name: "Test invite orga"
      })
    end
    org = Organisation.find_by(name: "Test invite orga")

    resp = OrganisationService.setup_payment(org)
    assert !resp.success
   end

  test "setup payment handle stripe error" do 
    StripeMock.prepare_error(StandardError.new("Test error"), :new_setup_intent)
    resp = OrganisationService.setup_payment(organisations(:kais_org))
    assert !resp.success
    assert_includes resp.message, "Test error"
   end

  test "default payment" do 
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id

    resp = OrganisationService.make_payment_method_default(org,"card_aaa")
    assert resp.success
   end

  test "default payment for org w/o customer" do 
    EventService.stub :new, @eventMock do
    OrganisationService.create_organisation_for_invite({
        name: "Test invite orga"
      })
    end
    org = Organisation.find_by(name: "Test invite orga")

    resp = OrganisationService.make_payment_method_default(org,"card_aaa")
    assert !resp.success
   end

  test "default payment handle stripe error" do 
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id

    StripeMock.prepare_error(StandardError.new("Test error"), :update_customer)
    resp = OrganisationService.make_payment_method_default(org,"card_aaa")
    assert !resp.success
    assert_includes resp.message, "Test error"
  end

  test "delete org" do 
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id

   EventService.stub :new, @eventMock do
      OrganisationService.delete(organisation: org, soft_delete: false)
    end

    assert_raises do
      Organisation.find(org.id)
    end
    assert_mock @eventMock
  end

  test "soft delete org" do
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id

    EventService.stub :new, @eventMock do
      OrganisationService.delete(organisation: org, soft_delete: true)
    end

    org.reload
    assert_equal "deleted", org.status
    assert_mock @eventMock
  end

  test "garbage collect orphaned users" do
    org = organisations(:kais_org)
    org.stripe_customer_id = Stripe::Customer.create({}).id
    org.subscription.stripe_subscription_id = Stripe::Subscription.create({
      customer: org.stripe_customer_id,
      items: [ { price: @price.id }, ],
    }).id

   EventService.stub :new, @eventMock do
      OrganisationService.delete(organisation: org, soft_delete: false)
    end

    assert_raises do
      User.find(users(:kai_notify1).id)
    end
    assert_raises do
      User.find(users(:kai_notify2).id)
    end
    assert User.find(users(:kai).id)
 end
end
