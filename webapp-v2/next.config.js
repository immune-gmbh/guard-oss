const nextTranslate = require('next-translate');

module.exports = nextTranslate({
  publicRuntimeConfig: {
    hosts: {
      frontend: process.env.FRONTEND_HOST,
      apiSrv: process.env.APISRV_HOST,
      authSrv: process.env.AUTHSRV_HOST,
      internalAuthSrv: process.env.INTERNAL_AUTHSRV_HOST,
    },
    developmentBearerToken: process.env.DEVELOPMENT_BEARER_TOKEN,
    stripeApiKey: process.env.STRIPE_API_KEY,
    isEdge: !!process.env.APPSRV_IS_EDGE,
    disableRegistration: !!process.env.APPSRV_DISABLE_REGISTRATION,
  },
});
