module RakeInstrumentation
  def self.included(base)
    base.prepend(Patch)
  end

  module Patch
    def execute(args = nil)
      attribs = {
        'rake.execute' => name,
        'rake.args' => args.inspect,
      }
      ret = nil
      GlobalTracer.in_span('Task#execute', attributes: attribs) do
        ret = super args
      end
    ensure
      OpenTelemetry.tracer_provider.force_flush
      ret
    end

    def invoke(*args)
      attribs = {
        'rake.invoke' => name,
        'rake.args' => args.inspect,
      }
      ret = nil
      GlobalTracer.in_span('Task#invoke', attributes: attribs) do
        ret = super args
      end
      ret
    end
  end
end

if Settings.telemetry.enabled
  endpoint = Settings.telemetry.endpoint
  spanExporter =
    case Settings.telemetry.protocol
    when "otlp"
      token = ENV.fetch 'AUTHSRV_OTLP_TOKEN' do '' end
      OpenTelemetry::Exporter::OTLP::Exporter.new(
        endpoint: endpoint,
        headers: {"lightstep-access-token" => token})
    when "jaeger"
      OpenTelemetry::Exporter::Jaeger::CollectorExporter.new(endpoint: endpoint)
    else
      raise ArgumentError
    end

  spanProc = OpenTelemetry::SDK::Trace::Export::BatchSpanProcessor.new(spanExporter)

  OpenTelemetry::SDK.configure do |c|
    c.use_all
    c.add_span_processor(spanProc)
    c.service_name = Settings.telemetry.name
    c.resource = OpenTelemetry::Resource::Detectors::AutoDetector.detect.merge(OpenTelemetry::SDK::Resources::Resource.create({
      OpenTelemetry::SemanticConventions::Resource::SERVICE_VERSION => Settings.release,
      OpenTelemetry::SemanticConventions::Resource::HOST_NAME => Socket.gethostname,
      "code.repository" => "https://github.com/immune-gmbh/guard/authsrv2",
    }))
  end

  ::Rake::Task.include(RakeInstrumentation)
end

GlobalTracer = OpenTelemetry.tracer_provider.tracer("authsrv", "v2")
