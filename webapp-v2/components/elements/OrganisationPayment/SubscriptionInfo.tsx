import Link from 'components/elements/Button/Link';
import Headline from 'components/elements/Headlines/Headline';
import InvoicesTable from 'components/elements/OrganisationPayment/InvoicesTable';
import { format } from 'date-fns';
import { useSubscription } from 'hooks/subscriptions';
import React from 'react';

const FeeStructure = ({
  baseFee,
  deviceFee,
  deviceNum,
}: {
  baseFee: number;
  deviceFee: number;
  deviceNum: number;
}): JSX.Element => {
  const deviceFeeSum = deviceFee * deviceNum;

  if (baseFee <= 0 && deviceFee <= 0) {
    return <p>This organisation is free.</p>;
  } else if (baseFee > 0 && deviceFee <= 0) {
    return <p>This organisation is billed a monthly flat fee of {baseFee}€.</p>;
  } else if (baseFee <= 0 && deviceFee > 0) {
    return (
      <p>
        This organisation is billed {deviceFee}€ per device i.e. {deviceFeeSum}€ per month.
      </p>
    );
  } /*if (baseFee > 0 && deviceFee > 0)*/ else {
    return (
      <p>
        This organisation is billed {deviceFee}€ per device ({deviceFeeSum}€) plus a base fee of{' '}
        {baseFee}€ per month.
      </p>
    );
  }
};

const CreditBalance = ({
  freeCredits,
  nextBill,
}: {
  freeCredits: number;
  nextBill: number;
}): JSX.Element => {
  if (freeCredits > 0 && nextBill <= 0) {
    return <p>This organisation has {freeCredits}€ of credit remaining.</p>;
  } else if (freeCredits > 0 && nextBill > 0) {
    if (freeCredits >= nextBill) {
      return (
        <p>
          The next bill will be covered by the organisation&apos;s credit balance of {freeCredits}€.
        </p>
      );
    } else {
      return (
        <p>
          The next bill will be covered partially by the organisation&apos;s credit balance of{' '}
          {freeCredits}€.
        </p>
      );
    }
  } else {
    return <p>No free credit remaining.</p>;
  }
};

interface ISubscriptionInfoProps {
  subscriptionId: string;
  showInvoices: boolean;
}

const SubscriptionInfo = ({
  subscriptionId,
  showInvoices,
}: ISubscriptionInfoProps): JSX.Element => {
  const { data: subscription } = useSubscription(subscriptionId);
  const freeCredits = subscription?.billingDetails?.freeCredits / 100;
  const nextBill =
    subscription?.monthlyBaseFee +
    subscription?.monthlyFeePerDevice * subscription?.currentDevicesAmount;

  return (
    <>
      {subscription?.periodEnd && (
        <p className="font-bold">
          Next Payment on {format(new Date(subscription?.periodEnd), 'PPP')}
        </p>
      )}
      {subscription?.currentDevicesAmount > 0 && (
        <p>This organisation is billed monthly for {subscription?.currentDevicesAmount} devices.</p>
      )}
      {subscription?.currentDevicesAmount == 0 && <p>No chargable devices.</p>}
      <FeeStructure
        baseFee={subscription?.monthlyBaseFee}
        deviceFee={subscription?.monthlyFeePerDevice}
        deviceNum={subscription?.currentDevicesAmount}
      />
      {freeCredits > 0 && <CreditBalance freeCredits={freeCredits} nextBill={nextBill} />}
      {showInvoices && (
        <>
          <Headline size={4}>Invoices</Headline>
          <InvoicesTable subscriptionId={subscriptionId} />
        </>
      )}
      <p>If you are interested in enterprise features, reach out to our sales department.</p>

      <div className="inline-block">
        <Link href="mailto:sales@immu.ne?subject=Device%20limit%20reached" target="_blank">
          Contact sales
        </Link>
      </div>
    </>
  );
};
export default SubscriptionInfo;
