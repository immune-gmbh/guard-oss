if Rails.env.development?
  Js::Routes::Rails.configure do |c|
    c.output = ENV.fetch 'AUTHSRV_JS_ROUTES_PATH' do
      '../webapp-v2/generated/authsrvRoutes.js' 
    end
    c.template = :commonjs
  end
end
