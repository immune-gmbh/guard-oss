# class BlahController < ApplicationController
#   include Traceable
# end
module Traceable
  extend ActiveSupport::Concern

  def trace_action
    span = OpenTelemetry::Trace.current_span
    if span&.context&.valid?
      span.set_attribute('code.file', __FILE__)
      span.set_attribute('actor', current_actor.inspect)
    end

    begin
      yield
    rescue StandardError => e
      span.record_exception e
      span.status = OpenTelemetry::Trace::Status.error(e.message)
      raise e
    ensure
      span.status = OpenTelemetry::Trace::Status.error("HTTP status code #{response.status}") if response.status >= 400
    end
  end

  included do
    around_action :trace_action

    def current_span
      OpenTelemetry::Trace.current_span
    end

    def trace(operation, &block)
      GlobalTracer.in_span operation, &block
    end

    def trace_exception(exception)
      logger.info "Exception: #{exception.inspect}"
      current_span.record_exception exception
    end

    def trace_event(event)
      logger.info event.inspect.to_s
      current_span.add_event event
    end
  end
end
