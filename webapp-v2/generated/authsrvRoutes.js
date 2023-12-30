const JsRoutesRails = (function() {
  var routes = {};
  
  routes['v2_session_path'] = function(options) {
    return format('/v2/session', options);
  };
  
  routes['activate_v2_user_path'] = function(options) {
    return format('/v2/users/:id/activate', options);
  };
  
  routes['resend_v2_user_path'] = function(options) {
    return format('/v2/users/:id/resend', options);
  };
  
  routes['change_password_v2_users_path'] = function(options) {
    return format('/v2/users/change_password', options);
  };
  
  routes['v2_users_path'] = function(options) {
    return format('/v2/users', options);
  };
  
  routes['new_v2_user_path'] = function(options) {
    return format('/v2/users/new', options);
  };
  
  routes['edit_v2_user_path'] = function(options) {
    return format('/v2/users/:id/edit', options);
  };
  
  routes['v2_user_path'] = function(options) {
    return format('/v2/users/:id', options);
  };
  
  routes['setup_v2_subscriptions_path'] = function(options) {
    return format('/v2/subscriptions/setup', options);
  };
  
  routes['download_v2_subscription_invoice_path'] = function(options) {
    return format('/v2/subscriptions/:subscription_id/invoices/:id/download', options);
  };
  
  routes['v2_subscription_invoices_path'] = function(options) {
    return format('/v2/subscriptions/:subscription_id/invoices', options);
  };
  
  routes['v2_subscription_invoice_path'] = function(options) {
    return format('/v2/subscriptions/:subscription_id/invoices/:id', options);
  };
  
  routes['v2_subscriptions_path'] = function(options) {
    return format('/v2/subscriptions', options);
  };
  
  routes['v2_subscription_path'] = function(options) {
    return format('/v2/subscriptions/:id', options);
  };
  
  routes['devices_v2_organisations_path'] = function(options) {
    return format('/v2/organisations/devices', options);
  };
  
  routes['v2_organisation_memberships_path'] = function(options) {
    return format('/v2/organisations/:organisation_id/memberships', options);
  };
  
  routes['new_v2_organisation_membership_path'] = function(options) {
    return format('/v2/organisations/:organisation_id/memberships/new', options);
  };
  
  routes['edit_v2_organisation_membership_path'] = function(options) {
    return format('/v2/organisations/:organisation_id/memberships/:id/edit', options);
  };
  
  routes['v2_organisation_membership_path'] = function(options) {
    return format('/v2/organisations/:organisation_id/memberships/:id', options);
  };
  
  routes['v2_organisations_path'] = function(options) {
    return format('/v2/organisations', options);
  };
  
  routes['new_v2_organisation_path'] = function(options) {
    return format('/v2/organisations/new', options);
  };
  
  routes['edit_v2_organisation_path'] = function(options) {
    return format('/v2/organisations/:id/edit', options);
  };
  
  routes['v2_organisation_path'] = function(options) {
    return format('/v2/organisations/:id', options);
  };
  
  routes['activate_v2_membership_path'] = function(options) {
    return format('/v2/memberships/:id/activate', options);
  };
  
  routes['v2_memberships_path'] = function(options) {
    return format('/v2/memberships', options);
  };
  
  routes['new_v2_membership_path'] = function(options) {
    return format('/v2/memberships/new', options);
  };
  
  routes['edit_v2_membership_path'] = function(options) {
    return format('/v2/memberships/:id/edit', options);
  };
  
  routes['v2_membership_path'] = function(options) {
    return format('/v2/memberships/:id', options);
  };
  
  routes['v2_password_reset_index_path'] = function(options) {
    return format('/v2/password_reset', options);
  };
  
  routes['edit_v2_password_reset_path'] = function(options) {
    return format('/v2/password_reset/:id/edit', options);
  };
  
  routes['v2_password_reset_path'] = function(options) {
    return format('/v2/password_reset/:id', options);
  };
  
  routes['v2_events_path'] = function(options) {
    return format('/v2/events', options);
  };
  
  routes['v2_healthy_path'] = function(options) {
    return format('/v2/healthy', options);
  };
  
  routes['v2_ready_path'] = function(options) {
    return format('/v2/ready', options);
  };
  
  routes['v2_stripe_webhook_path'] = function(options) {
    return format('/v2/stripe_webhooks', options);
  };
  
  routes['v2_path'] = function(options) {
    return format('/v2/oauth/callback/:provider', options);
  };
  
  routes['v2_auth_at_provider_path'] = function(options) {
    return format('/v2/oauth/:provider', options);
  };
  

  function format(string, options) {
    var str = string.toString();
    for (var option in options) {
      str = str.replace(RegExp("\\:" + option, "gi"), options[option]);
    }

    return str;
  };

  return routes;
})();

export { JsRoutesRails };
