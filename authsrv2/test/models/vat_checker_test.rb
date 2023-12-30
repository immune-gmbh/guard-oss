require "test_helper"

class VatCheckerTest < ActiveSupport::TestCase
  test "stripe tax id" do
    assert_equal "eu_vat", VatChecker.stripe_taxid_for_country("Germany", "DE123456789")
    assert_equal "eu_vat", VatChecker.stripe_taxid_for_country("Finland", "FI12345678")
    assert_equal "gb_vat", VatChecker.stripe_taxid_for_country("United Kingdom", "GB123456789")
    assert_equal "eu_vat", VatChecker.stripe_taxid_for_country("United Kingdom", "XI123456789")
    assert_equal "ru_inn", VatChecker.stripe_taxid_for_country("Russia", "1234567891")
    assert_equal "ru_kpp", VatChecker.stripe_taxid_for_country("Russia", "123456789")
  end

  test "invalid stripe taxids" do
    assert_raises ArgumentError do
      VatChecker.stripe_taxid_for_country("Atlantis", "12345678")
    end

    assert_raises ArgumentError do
      VatChecker.stripe_taxid_for_country("Russia", "123")
    end
  end
end
