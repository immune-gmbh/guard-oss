module V2
  class ConfigurationsController < V2::ApiBaseController
    def show
      render json: {
        config: {
          release: Settings.release || '0',
          agent_urls: {
            ubuntu: Settings.agent_urls&.ubuntu || 'not/available',
            fedora: Settings.agent_urls&.fedora || 'not/available',
            generic: Settings.agent_urls&.generic || 'not/available',
            windows: Settings.agent_urls&.windows || 'not/available',
          }
        }
      }
    end
  end
end
