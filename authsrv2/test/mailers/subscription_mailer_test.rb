require "test_helper"

class SubscriptionMailerTest < ActionMailer::TestCase
  setup do
    @membership = memberships(:kai_at_kais_org)
    @subscription = subscriptions(:kais_subscription)
  end

  test "new freeloader sub" do
    @membership.organisation.freeloader = true
    @membership.organisation.save!
    email = SubscriptionMailer.created_email(@subscription)

    assert_emails 1 do
      email.deliver_now
    end

    assert_equal ["support@immune.gmbh"], email.from
    assert_equal ["kai.michaelis@immu.ne"], email.to
    assert_match /New Subscription/, email.subject
    assert_match /As your organisation is running on our free plan/, email.body.to_s
  end

  test "new paid sub" do
    email = SubscriptionMailer.created_email(@subscription)

    assert_emails 1 do
      email.deliver_now
    end

    assert_equal ["support@immune.gmbh"], email.from
    assert_equal ["kai.michaelis@immu.ne"], email.to
    assert_match /New Subscription/, email.subject
    assert_match /After the initial month, you'll be charged/, email.body.to_s
  end
end
