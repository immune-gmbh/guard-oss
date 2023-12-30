import * as IssuesV1 from 'generated/issuesv1';
import { ApiSrv } from 'types/apiSrv';
import { convertToCamelRecursive } from 'utils/case';
import { deserialiseWithoutMeta } from 'utils/session';
import postSessionFixture from 'mocks/fixtures/post-session.json';
import getOrgsFx from 'mocks/fixtures/get-organisations.json';
import getOrgsA3Fx from 'mocks/fixtures/get-organisations-a3e85ce4-7687-5078-8716-c6c3dfaf554f.json'
import getOrgsA3UsersFx from 'mocks/fixtures/get-organisations-a3e85ce4-7687-5078-8716-c6c3dfaf554f-include[]=users.json'
import getUsersWithOrgsFx from 'mocks/fixtures/get-users-include-orgs.json';
import getMembershipsFx from 'mocks/fixtures/get-memberships.json';
import getMembership628Fx from 'mocks/fixtures/get-memberships-628cf88a-94a9-55bf-bea4-fea7e6f4f4bc.json';
import getSubscriptionInvoices23Fx from 'mocks/fixtures/get-subscriptions-233eca70-6419-5d2b-8006-1b6a6392046b-invoices.json';
import getSubscription23Fx from 'mocks/fixtures/get-subscriptions-233eca70-6419-5d2b-8006-1b6a6392046b.json';
import getDevice1770Fx from 'mocks/fixtures/get-devices-1770.json';
import getDevicesFx from 'mocks/fixtures/devices.json';
import { SerializedInvoice, SerializedMembership, SerializedOrganisation, SerializedSession, SerializedSubscription, SerializedUser } from 'utils/types';

export const tagsMock = [
  {
    id: 'mzb5XFl3m',
    key: 'Blah',
  },
  {
    id: '20vlEOys9E',
    key: 'Blub',
  },
  {
    id: '2AWyQQrM',
    key: 'Foo',
  },
] as ApiSrv.Tag[];

export const devicesMock = [
  {
    id: '1770',
    tags: tagsMock,
    hwid: 'af0bff38ba3f50d0716e505ac1624af2e',
    name: 'fedora',
    state: 'outdated',
  },
  {
    id: '1772',
    tags: [tagsMock[0]],
    hwid: '0022000b74795e64b75f5694800e30e7568fd7',
    name: 'ludmilla',
    state: 'trusted',
  },
] as unknown as ApiSrv.Device[];

export const mockDeviceStats = {
  numTrusted: 102,
  numWithIncident: 4,
  numAtRisk: 0,
  numUnresponsive: 88,
};

export const mockRisks = [
  { issueId: 'csme/no-update', count: 2 },
  { issueId: 'uefi/no-exit-boot-srv', count: 10 },
  { issueId: '"uefi/secure-boot-keys', count: 1 },
  { issueId: 'uefi/boot-order', count: 5 },
  { issueId: 'uefi/option-rom-set', count: 4 },
  { issueId: 'uefi/secure-boot-dbx', count: 3 },
  { issueId: 'uefi/ibb-no-update', count: 1 },
];

export const mockIncidentList = [
  {
    issueId: 'tpm/endorsement-cert-unverified',
    timestamp: '2023-02-15T10:41:42.865613Z',
    count: 8,
  },
  {
    issueId: 'tpm/invalid-eventlog',
    timestamp: '2023-02-15T10:41:42.865613Z',
    count: 7,
  },
];

export const mockRiskList = [
  {
    issueId: 'tpm/no-eventlog',
    count: 10,
  },
  {
    issueId: 'uefi/official-dbx',
    count: 22,
  },
];

export const mockDashboardData = {
  deviceStats: mockDeviceStats,
  incidents: {
    count: 10,
    devices: 30,
  },
  risks: [mockRisks[0]],
};

export const mockSessionContext = {
  session: {
    user: { id: '1', role: 'user' },
    currentMembership: {
      organisation: {
        id: '2',
        name: 'My Organisation'
      },
    },
  },
  isInitialized: true,
  setSession: () => {}
};

export const mockConfig = {
  config: {
    release: 'test env',
    agent_urls: {},
  }
}

export const mockSubscription = {
  monthlyBaseFee: 15,
  monthlyFeePerDevice: 5,
  currentDevicesAmount: 50,
  billingDetails: {
    freeCredits: 0,
  }
} as SerializedSubscription

export const mockReportCardIncident = ({
  id,
  aspect = 'firmware',
  args,
}: {
  id: string;
  aspect?: string;
  args?: any;
}) =>
  ({
    id,
    incident: true,
    aspect,
    args: { ...args },
  } as IssuesV1.HttpsImmuneAppSchemasIssuesv1SchemaYaml);


export const serializedSession = deserialiseWithoutMeta(
  convertToCamelRecursive(postSessionFixture)
) as SerializedSession;

export const serializedOrganisations = deserialiseWithoutMeta(
  convertToCamelRecursive(getOrgsFx)
) as SerializedOrganisation[];


export const serializedOrganisation = deserialiseWithoutMeta(
  convertToCamelRecursive(getOrgsA3Fx)
) as SerializedOrganisation;

export const serializedOrganisationWithUsers = deserialiseWithoutMeta(
  convertToCamelRecursive(getOrgsA3UsersFx)
) as SerializedOrganisation;

export const serializedUsersWithOrgs = deserialiseWithoutMeta(
  convertToCamelRecursive(getUsersWithOrgsFx)
) as SerializedUser[];

export const serializedMemberships = deserialiseWithoutMeta(
  convertToCamelRecursive(getMembershipsFx)
) as SerializedMembership[];

export const serializedMembership = deserialiseWithoutMeta(
  convertToCamelRecursive(getMembership628Fx)
) as SerializedMembership;

export const serializedSubscription = deserialiseWithoutMeta(
  convertToCamelRecursive(getSubscription23Fx)
) as SerializedSubscription;

export const serializedSubscriptionInvoices = deserialiseWithoutMeta(
  convertToCamelRecursive(getSubscriptionInvoices23Fx)
) as SerializedInvoice[];

export const serializedDeviceWithAppraisals = deserialiseWithoutMeta(
  convertToCamelRecursive(getDevice1770Fx)
) as ApiSrv.Device;

export const serializedDevices = deserialiseWithoutMeta(
  convertToCamelRecursive(getDevicesFx)
) as ApiSrv.Device[];

export const mockElementFn = () => ({
  mount: jest.fn(),
  destroy: jest.fn(),
  on: jest.fn(),
  update: jest.fn(),
})

export const mockElementsFn = () => {
  const elements = {};
  return {
    create: jest.fn((type: string) => {
      elements[type] = mockElementFn();
      return elements[type];
    }),
    getElement: jest.fn((type: string) => {
      return elements[type] || null;
    }),
  };
};

export const mockStripeFn = () => ({
  elements: jest.fn(() => mockElementsFn()),
  createToken: jest.fn(),
  createSource: jest.fn(),
  createPaymentMethod: jest.fn(),
  confirmCardPayment: jest.fn(),
  confirmCardSetup: jest.fn(),
  paymentRequest: jest.fn(),
  _registerWrapper: jest.fn(),
});
