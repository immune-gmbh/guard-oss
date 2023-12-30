class Hash
  def to_o
    JSON.parse to_json, object_class: OpenStruct
  end
end

class Appraisal
  COMPONENTS_V1 = /^(bootchain|configuration|firmware)_ok\?$/
  COMPONENTS_V2 = /^(supply_chain|bootloader|operating_system|endpoint_protection)_ok\?$/

  def initialize(json)
    @raw = JSON.parse json
  end

  def method_missing(meth)
    if meth.id2name =~ COMPONENTS_V1 || meth.id2name =~ COMPONENTS_V2
      verdict_component meth.id2name
    else
      @raw[meth.id2name]
    end
  end

  def bootchain_ok?
    v = @raw['Verdict'] || @raw['verdict']
    if v.is_a?(Hash)
      v['bootchain']
    else
      report&.invalid_pcrs&.empty?
    end
  end

  def received
    DateTime.strptime(@raw['received'], '%Y-%m-%dT%H:%M:%S.%N%z')
  rescue StandardError
    DateTime.strptime(@raw['received'], '%Y-%m-%dT%H:%M:%S%z')
  end

  def report
    @raw['report'].to_o
  end

  def verdict
    v = @raw['Verdict'] || @raw['verdict']
    case v
    when true, false
      v
    when Hash
      case v['type']
      when 'verdict/3'
        v['result'] == 'trusted'
      when 'verdict/2', 'verdict/1'
        v['result']
      else
        raise ArgumentError
      end
    else
      raise ArgumentError
    end
  end

  private

  def verdict_component(comp)
    v = @raw['Verdict'] || @raw['verdict']
    if v.is_a?(Hash)
      if comp =~ COMPONENTS_V2 && v['type'] == 'verdict/3'
        v[comp.gsub(/_ok\?$/, '')] != 'vulnerable'
      elsif comp =~ COMPONENTS_V2 && v['type'] == 'verdict/2'
        v[comp.gsub(/_ok\?$/, '')]
      elsif comp == 'bootchain_ok' && (v['type'] == 'verdict/2' || v['type'] == 'verdict/3')
        true
      elsif comp =~ COMPONENTS_V1
        v[comp.gsub(/_ok\?$/, '')]
      else
        true
      end
    else
      true
    end
  end
end
