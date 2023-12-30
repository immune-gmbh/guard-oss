import getAppconfigFx from 'mocks/fixtures/get-appconfig.json';
import getDevicesFx from 'mocks/fixtures/devices.json';
import getChangesFx from 'mocks/fixtures/get-changes.json';
import getDeviceAppraisals1770Fx from 'mocks/fixtures/get-devices-1770-appraisals.json';
import getDevice1770Fx from 'mocks/fixtures/get-devices-1770.json';
import getOrgsA3UsersFx from 'mocks/fixtures/get-organisations-a3e85ce4-7687-5078-8716-c6c3dfaf554f-include[]=users.json';
import getOrgsFx from 'mocks/fixtures/get-organisations.json';
import getSubscription23Fx from 'mocks/fixtures/get-subscriptions-233eca70-6419-5d2b-8006-1b6a6392046b.json';
import getSubscriptionInvoices23Fx from 'mocks/fixtures/get-subscriptions-233eca70-6419-5d2b-8006-1b6a6392046b-invoices.json';
import getUser30Fx from 'mocks/fixtures/get-users-307f78d7-877b-53d0-abb3-c063159413fc.json';
import getUsersFx from 'mocks/fixtures/get-users.json';
import getUsersWithOrgsFx from 'mocks/fixtures/get-users-include-orgs.json';
import getDashboardFx from 'mocks/fixtures/get-dashboard.json';
import getRisksFx from 'mocks/fixtures/get-risks.json';
import getTagsFx from 'mocks/fixtures/get-tags.json';
import getIncidentsFx from 'mocks/fixtures/get-incidents.json';
import getMembershipsFx from 'mocks/fixtures/get-memberships.json';
import getMembership628Fx from 'mocks/fixtures/get-memberships-628cf88a-94a9-55bf-bea4-fea7e6f4f4bc.json';
import postSessionFx from 'mocks/fixtures/post-session.json';
import postSubscriptionsIntentFx from 'mocks/fixtures/post-subscriptions-intent.json';
import { rest } from 'msw';

export const handlers = [
  // authsrv
  rest.post('http://localhost:3000/v2/session', (_req, res, ctx) => {
    postSessionFx.data.attributes.next_path = 'http://localhost:8080/dashboard/welcome';
    return res(ctx.json(postSessionFx));
  }),
  rest.get('http://localhost:3000/v2/session/refresh', (_req, res, ctx) => {
    postSessionFx.data.attributes.next_path = 'http://localhost:8080/dashboard/welcome';
    return res(ctx.json(postSessionFx));
  }),
  rest.get('http://localhost:3000/v2/appconfig', (_req, res, ctx) => {
    return res(ctx.json(getAppconfigFx));
  }),
  rest.get(
    'http://localhost:3000/v2/users/307f78d7-877b-53d0-abb3-c063159413fc',
    (_req, res, ctx) => {
      return res(ctx.json(getUser30Fx));
    },
  ),
  rest.patch(
    'http://localhost:3000/v2/users/307f78d7-877b-53d0-abb3-c063159413fc',
    (_req, res, ctx) => {
      return res(ctx.json(getUser30Fx));
    },
  ),
  rest.get('http://localhost:3000/v2/users', (req, res, ctx) => {
    const withOrganisations = req.url.searchParams.get('include') === 'organisations'

    if (withOrganisations) return res(ctx.json(getUsersWithOrgsFx));

    return res(ctx.json(getUsersFx));
  }),
  rest.get('http://localhost:3000/v2/organisations', (_req, res, ctx) => {
    return res(ctx.json(getOrgsFx));
  }),
  rest.post('http://localhost:3000/v2/organisations', (_req, res, ctx) => {
    return res(ctx.json(getOrgsFx));
  }),
  rest.get(
    'http://localhost:3000/v2/organisations/a3e85ce4-7687-5078-8716-c6c3dfaf554f',
    (_req, res, ctx) => {
      return res(ctx.json(getOrgsA3UsersFx));
    },
  ),
  rest.patch(
    'http://localhost:3000/v2/organisations/a3e85ce4-7687-5078-8716-c6c3dfaf554f',
    (_req, res, ctx) => {
      return res(ctx.json(getOrgsA3UsersFx));
    },
  ),
  rest.get(
    'http://localhost:3000/v2/subscriptions/233eca70-6419-5d2b-8006-1b6a6392046b',
    (_req, res, ctx) => {
      return res(ctx.json(getSubscription23Fx));
    },
  ),
  rest.get(
    'http://localhost:3000/v2/subscriptions/233eca70-6419-5d2b-8006-1b6a6392046b/invoices',
    (_req, res, ctx) => {
      return res(ctx.json(getSubscriptionInvoices23Fx));
    },
  ),
  rest.post('http://localhost:3000/v2/subscriptions/intent', (_req, res, ctx) => {
    return res(ctx.json(postSubscriptionsIntentFx));
  }),
  rest.get('http://localhost:3000/v2/memberships', (req, res, ctx) => {
    return res(ctx.json(getMembershipsFx));
  }),
  rest.patch('http://localhost:3000/v2/memberships/628cf88a-94a9-55bf-bea4-fea7e6f4f4bc', (req, res, ctx) => {
    return res(ctx.json(getMembership628Fx));
  }),

  // apisrv
  rest.get('http://localhost:9292/v2/tags', (req, res, ctx) => {
    return res(ctx.json(getTagsFx));
  }),
  rest.get('http://localhost:9292/v2/devices', (req, res, ctx) => {
    if (req.url.searchParams.has('i')) {
      const cpy = Object.assign({}, getDevicesFx);
      cpy.data = [];
      return res(ctx.json(cpy));
    }
    return res(ctx.json(getDevicesFx));
  }),
  rest.get('http://localhost:9292/v2/devices/1770', (_req, res, ctx) => {
    return res(ctx.json(getDevice1770Fx));
  }),
  rest.get('http://localhost:9292/v2/devices/1770/appraisals', (_req, res, ctx) => {
    return res(ctx.json(getDeviceAppraisals1770Fx));
  }),
  rest.get('http://localhost:9292/v2/changes', (_req, res, ctx) => {
    return res(ctx.json(getChangesFx));
  }),
  rest.get('http://localhost:9292/v2/dashboard', (req, res, ctx) => {
    return res(ctx.json(getDashboardFx));
  }),
  rest.get('http://localhost:9292/v2/risks', (req, res, ctx) => {
    return res(ctx.json(getRisksFx));
  }),
  rest.get('http://localhost:9292/v2/incidents', (req, res, ctx) => {
    return res(ctx.json(getIncidentsFx));
  }),
];
