# Alerts are generated from failed appraisals and transformed into various
# formats like emails, SIEM alerts and log entries.
class Alert
  attr_accessor :description, :title, :id, :appraisal, :hostname, :timestamp, :device, :name

  def self.failed_appraisal(device, appraisal)
    return [] if appraisal.verdict

    ret = []
    device = OpenStruct.new(device)

    # deprecated
    unless appraisal.bootchain_ok?
      a = Alert.new(device, appraisal)
      a.id = 'pcr-changed'
      a.title = 'Firmware manipulated'

      if appraisal&.invalid_pcrs&.size == 1
        f = appraisal.invalid_pcrs.entries.first
        a.description = "#{Alert.friendly_pcr_name f[0].to_i} was changed. This indicates an unauthorized reconfiguration or malware."
      elsif (num_pcr = appraisal&.invalid_pcrs&.size) && num_pcr.positive?
        a.description = "Firmware code and data were changed in #{num_pcr} areas. This indicates an unauthorized reconfiguration or malware."
      else
        a.description = 'Firmware code and data were manipulated. This indicates an unauthorized reconfiguration or malware.'
      end

      ret << a
    end

    unless appraisal.supply_chain_ok?
      a = Alert.new(device, appraisal)
      a.id = 'supply-chain'
      a.title = 'Supply chain issue'
      a.description = 'The supply chain of the device could not be verified.'

      ret << a
    end

    unless appraisal.configuration_ok?
      a = Alert.new(device, appraisal)
      a.id = 'insecure-configuration'
      a.title = 'Device configuration insecure'
      a.description = 'The firmware configuration of this device is insecure.'

      ret << a
    end

    unless appraisal.firmware_ok?
      a = Alert.new(device, appraisal)
      a.id = 'firmware-vulnerable'
      a.title = 'Firmware vulnerable'
      a.description = 'The firmware on this device has known security vulnerabilities.'

      ret << a
    end

    unless appraisal.bootloader_ok?
      a = Alert.new(device, appraisal)
      a.id = 'bootloader-vulnerable'
      a.title = 'Bootloader configuration manipulated'
      a.description = "The a problem with the device's bootloader configuration was found."

      ret << a
    end

    unless appraisal.operating_system_ok?
      a = Alert.new(device, appraisal)
      a.id = 'operating-system-vulnerable'
      a.title = 'Operating system manipulated'
      a.description = "The a problem with the device's operating system was found."

      ret << a
    end

    unless appraisal.endpoint_protection_ok?
      a = Alert.new(device, appraisal)
      a.id = 'endpoint-protection-insecure'
      a.title = 'Endpoint protection issue'
      a.description = "The a problem with the device's endpoint protection solution was found."

      ret << a
    end

    # no specific failure. generate a generic (heh) alert
    # deprecated
    if ret.empty?
      a = Alert.new(device, appraisal)
      a.id = 'attestation-failed'
      a.title = 'Attestation failed'
      a.description = 'The attestation of the device failed due to an unknown reason.'
      ret << a
    end

    ret
  end

  def self.expired_appraisal(device, appraisal)
    device = OpenStruct.new(device)

    a = Alert.new(device, appraisal)
    a.id = 'appraisal-expired'
    a.title = 'Attestation out of date'
    a.description = 'The device failed to send a fresh attestation.'
    [a]
  end

  private

  def initialize(device, appraisal)
    @appraisal = appraisal.id.to_s
    @timestamp = appraisal.received
    @hostname = (appraisal.report || {})['hostname'] || ''
    @device = device.id
    @name = device.name
  end

  def self.friendly_pcr_name(pcr)
    case pcr
    when 0 then 'Initial boot block'
    when 1 then 'Host platform configuration'
    when 2 then 'UEFI driver and application code'
    when 3 then 'UEFI driver and application configuration and data'
    when 4 then 'UEFI Boot Manager code and boot attempts'
    when 5 then 'Boot Manager code configuration and data and GPT/partition table'
    when 6 then 'Host platform manufacturer specific code and data'
    when 7 then 'Secure Boot policy'
    when 8 then 'Operating system'
    when 9 then 'Operating system'
    when 10 then 'Operating system'
    when 11 then 'Operating system'
    when 12 then 'Operating system'
    when 13 then 'Operating system'
    when 14 then 'Operating system'
    when 15 then 'Operating system'
    when 16 then 'Debug PCR'
    when 17 then 'Secure operating system loader'
    when 18 then 'Secure boot loader and operating system'
    when 22 then 'Geolocation or asset tag'
    else "Platform configuration register ##{pcr}"
    end
  end
end
