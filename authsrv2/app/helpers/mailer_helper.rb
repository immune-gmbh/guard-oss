module MailerHelper
  def humanize_list(list, thing="", things="")
    if list.size == 1
      list[0] || ""
    elsif list.size < 3
      "#{list[0..-2].join(", ")} and #{list.last}"
    elsif list.size == 3
      "#{list[0..1].join(", ")} and #{list.size - 2} more#{thing}"
    else
      "#{list[0..1].join(", ")} and #{list.size - 2} more#{things}"
    end
  end

  def device_url(device, org)
    "#{Settings.external_frontend_url}/dashboard/devices/#{device}?organisation=#{org}"
  end

  def vulnerable_devices_url(org)
    "#{Settings.external_frontend_url}/dashboard/devices?state=vulnerable&organisation=#{org}"
  end
end
