class BaseSerializer
  include JSONAPI::Serializer

  meta do ||
    {
      version: Settings.release,
      instance: Socket.gethostname,
      trace: OpenTelemetry::Trace.current_span&.context&.hex_trace_id || ""
    }
  end

  def self.abilities(*actions)
    cancan_actions = self.expand_cancan_actions(actions)
    cancan_actions.each do |action|
      method = "can_#{action}".to_sym
      unless method_defined?(method)
        define_method method do |current_ability|
          current_ability.can? action, object
        end
      end

      send(:attributes, method, Proc.new do |object, params|
        params[:current_ability] && params[:current_ability].can?(action, object)
      end
      )
    end
  end

  def self.attribute_names
    []
  end

  private
  def self.expand_cancan_actions(actions)
    if actions.include? :restful
      actions.delete :restful
      actions |= [:index, :show, :new, :create, :edit, :update, :destroy]
    end
    actions
  end
end
