require 'kubeclient'
require 'concurrent'
require 'concurrent-edge'

class KeyDiscoveryService
  class << self
    attr_accessor :ping_channel, :background_thr, :client, :keyset_ref

    def keyset
      spawn_key_discovery
      keyset_ref&.get || {}
    end

    def keyset=(keyset)
      keyset_ref.set(keyset)
    end

    def ping
      spawn_key_discovery
      thr = Thread.new do
        ping_channel.put :ping
      end

      # we give the background thread 2s to respond
      thr.join 2
      st = thr.status
      thr.kill

      st == false
    end

    def spawn_key_discovery
      return if background_thr

      ticker = Concurrent::Channel.tick 3

      self.background_thr = Thread.new do
        loop do
          Concurrent::Channel.select do |s|
            # every 3s
            s.take ticker do
              GlobalTracer.in_span('kubernetesKeyDiscovery', kind: :client) do |span|
                newkeyset = discover_keys(span)
                span.add_event "#{newkeyset.size} keys discovered"
                span.set_attribute('keyset', newkeyset.inspect)

                self.keyset = newkeyset if Rails.env.production?
              end
            end

            # receiver of ping!
            s.take ping_channel do
            end

            s.default do
              sleep 0.1
            end
          end
        end
      end
    end

    def discover_keys(span)
      return [] if Rails.env.test? || Rails.env.development?

      new_client
      selector = Settings.authentication.label_selector
      new_keyset = {}

      self.client.get_pods(label_selector: selector).each do |pod|
        span&.add_event "check pod #{pod}"
        service_name = pod.metadata.labels['app.kubernetes.io/instance']
        next unless service_name

        span&.add_event "it's service #{service_name}"
        new_keyset[service_name] ||= []
        new_keyset[service_name] += extract_keys(pod, span)
      end

      new_keyset
    end

    def extract_keys(pod, span)
      pod.metadata.annotations.to_h.flat_map do |kv|
        next [] unless kv[0].start_with? 'immu.ne/public-key-'

        span&.add_event "got annotation #{kv[0]}=#{kv[1]}"

        begin
          ec = OpenSSL::PKey::EC.new(Base64.decode64(kv[1]))
          ec.check_key
          span&.set_attribute(kv[0].to_s, ec.inspect)
          [ec]
        rescue OpenSSL::PKey::PKeyError, ArgumentError => e
          span&.add_event("Error while processing pod #{pod.name} #{e.message}")
          []
        end
      end
    end

    def new_client
      return if client

      auth_options = {
        bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token'
      }
      ssl_options = {}
      if File.exist?('/var/run/secrets/kubernetes.io/serviceaccount/ca.crt')
        ssl_options[:ca_file] = '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
      end
      self.client = Kubeclient::Client.new(
        'https://kubernetes.default.svc',
        'v1',
        auth_options: auth_options,
        ssl_options: ssl_options
      )
    end
  end
end

KeyDiscoveryService.ping_channel = Concurrent::Channel.new
KeyDiscoveryService.keyset_ref = Concurrent::AtomicReference.new {}
