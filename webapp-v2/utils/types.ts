import schema from 'generated/authsrvSchema';

export interface Err {
  id?: string;
  title?: string;
}
export type Response<T> = T & {
  code: number;
  errors?: Err[];
  meta?: Record<string, unknown>;
};
export interface ActiveRecord {
  errors?: Err[];
  type: string;
  id: string;
  canRead?: boolean;
  canEdit?: boolean;
  canDelete?: boolean;
}

export interface SerializedUser
  extends ActiveRecord,
    Pick<schema.User, 'name' | 'email' | 'invited' | 'activationState'> {
  role: 'admin' | 'user';
  hasSeenIntro: boolean;
  hasPassword: boolean;
  authenticatedGithub: boolean;
  authenticatedGoogle: boolean;
  organisations: SerializedOrganisation[];
  subscriptions: SerializedSubscription[];
  address: SerializedAddress;
}

export interface SerializedSubscription
  extends ActiveRecord,
    Pick<
      schema.Subscription,
      'status' | 'currentDevicesAmount' | 'periodStart' | 'periodEnd' | 'taxRate'
    > {
  maxDevicesAmount: number;
  monthlyBaseFee: number;
  monthlyFeePerDevice: number;
  actionRequired: boolean;
  billingDetails: { last4: string; expiryDate: string; freeCredits: number };
}

export interface SerializedOrganisation extends ActiveRecord, Pick<schema.Organisation, 'name'> {
  memberships?: SerializedMembership[];
  memberCount: number;
  address?: SerializedAddress;
  vatNumber: string;
  subscription: SerializedSubscription;
  splunkEnabled: boolean;
  splunkEventCollectorUrl: string;
  splunkAcceptAllServerCertificates: boolean;
  splunkAuthenticationToken: string;
  syslogEnabled: boolean;
  syslogHostnameOrAddress: string;
  syslogUdpPort: string;
  invoiceName: string;
  freeloader: boolean;
  users?: SerializedUser[];
}
export interface SerializedMembership extends ActiveRecord {
  user: SerializedUser;
  notifyDeviceUpdate: boolean;
  notifyInvoice: boolean;
  role: 'owner' | 'admin' | 'user';
  organisation?: SerializedOrganisation;
  token?: string;
  enrollmentToken?: string;
}
export interface SerializedSession extends ActiveRecord {
  nextPath: string;
  defaultOrganisation: string;
  user: SerializedUser;
  memberships: SerializedMembership[];
}

export interface SerializedInvoice
  extends ActiveRecord,
    Pick<schema.Invoice, 'total' | 'status' | 'stripeInvoiceNumber' | 'finalizedAt'> {}

export interface SerializedAddress
  extends ActiveRecord,
    Pick<schema.Address, 'streetAndNumber' | 'city' | 'postalCode' | 'country'> {}

export type TrustedStates = 'VALID' | 'IGNORED' | 'VULNERABLE';
