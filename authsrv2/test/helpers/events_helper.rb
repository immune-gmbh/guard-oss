module EventsHelper
  def new_event(type: "ne.immu.v2.new-appraisal", subject: "test", data: {})
    event = CloudEvents::Event.create \
      spec_version:      "1.0",
      id:                "1234-1234-1234",
      source:            "/mycontext",
      type:              type,
      data_content_type: "application/json",
      data:              data,
      subject:           subject

    cloud_events_http = CloudEvents::HttpBinding.default
    cloud_events_http.encode_event event
  end

  def unknown_event
    new_event type: "com.example.hello-world", data: { message: "Hello, World" }
  end

  def new_appraisal(subject: nil, verdict: true, previous: nil)
    new_event type: "ne.immu.v2.new-appraisal",
      subject: subject || organisations(:kais_org).public_id,
      data: {
        next: { verdict: verdict },
        previous: previous,
        device: {}
      }
  end

  def expired_appraisal(subject: nil)
    new_event type: "ne.immu.v2.appraisal-expired",
      subject: subject || organisations(:kais_org).public_id,
      data: {}
  end

  def heartbeat(update_usage_records: true)
    new_event type: "ne.immu.v2.heartbeat",
      data: {
        update_usage_records: update_usage_records
      }
  end
end
