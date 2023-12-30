require "test_helper"

class AppraisalTest < ActiveSupport::TestCase
  test "boolean verdict" do
    body = JSON.load(file_fixture("failed-appraisal.json").read)
    appr = Appraisal.new(JSON.dump(body["data"]["next"]))

    assert_not appr.verdict
  end

  test "object verdict -- bootchain" do
    body = JSON.load(file_fixture("failed-appraisal-v2.json").read)
    appr = Appraisal.new(JSON.dump(body["data"]["next"]))

    assert_not appr.verdict
    assert_not appr.firmware_ok?
    assert appr.bootchain_ok?
    assert appr.configuration_ok?
  end

  test "object verdict -- firmware" do
    body = JSON.load(file_fixture("failed-appraisal-v3.json").read)
    appr = Appraisal.new(JSON.dump(body["data"]["next"]))

    assert_not appr.verdict
    assert_not appr.firmware_ok?
    assert appr.bootchain_ok?
    assert appr.configuration_ok?
  end

  test "object verdict -- configuration" do
    body = JSON.load(file_fixture("failed-appraisal-v4.json").read)
    appr = Appraisal.new(JSON.dump(body["data"]["next"]))

    assert_not appr.verdict
    assert appr.firmware_ok?
    assert appr.bootchain_ok?
    assert_not appr.configuration_ok?
  end

  test "good v2 appraisal" do
    body = JSON.load(file_fixture("good.appraisal.json").read)
    appr = Appraisal.new(JSON.dump(body))

    assert appr.verdict
    assert appr.supply_chain_ok?
    assert appr.configuration_ok?
    assert appr.firmware_ok?
    assert appr.bootloader_ok?
    assert appr.operating_system_ok?
    assert appr.endpoint_protection_ok?
    assert appr.bootchain_ok?
  end

  test "bad v2 appraisal" do
    body = JSON.load(file_fixture("no-secureboot.appraisal.json").read)
    appr = Appraisal.new(JSON.dump(body))

    assert_not appr.verdict
    assert appr.supply_chain_ok?
    assert_not appr.configuration_ok?
    assert appr.firmware_ok?
    assert appr.bootloader_ok?
    assert appr.operating_system_ok?
    assert appr.endpoint_protection_ok?
    assert appr.bootchain_ok?
  end
end
