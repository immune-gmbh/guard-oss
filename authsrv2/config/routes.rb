Rails.application.routes.draw do
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html
  namespace :v2, defaults: {format: :json}, export: true do
    resource :session, only: [:show, :create, :destroy] do
      get :refresh, on: :collection
    end

    resources :users do
      post :activate, on: :member
      post :resend, on: :member
      patch :change_password, on: :collection
    end

    resources :subscriptions, only: [:index, :create, :show] do
      collection do
        post :intent
        post :default_payment_method
      end

      resources :invoices, only: [:index, :show] do
        get :download, on: :member
      end
    end
    resources :organisations do
      patch :devices, on: :collection

      resources :memberships
    end
    resources :memberships do
      match :activate, on: :member, via: [:get, :post]
    end
    resources :password_reset, only: [:create, :edit, :update]

    # Events
    post '/events', to: 'events#receive'

    # K8s probes
    get '/healthy', to: 'probes#healthy'
    get '/ready', to: 'probes#ready'

    # Info Endpoints
    get '/appconfig', to: 'configurations#show'

    post 'stripe_webhooks' => 'stripe_webhooks#event', as: :stripe_webhook

    post 'oauth/callback/:provider' => 'oauths#callback'
    get 'oauth/callback/:provider' => 'oauths#callback' # for use with Github, Facebook
    get 'oauth/:provider' => 'oauths#oauth', :as => :auth_at_provider
  end

  mount LetterOpenerWeb::Engine, at: '/letter_opener' if Rails.env.development?
end
