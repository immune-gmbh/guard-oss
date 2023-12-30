import { SerializedSubscription } from 'utils/types';

interface ISubscriptionSummaryProps {
  freeloader: boolean;
  subscription: SerializedSubscription;
}

const SubscriptionSummary = ({
  freeloader,
  subscription,
}: ISubscriptionSummaryProps): JSX.Element => {
  const nextBill =
    subscription.monthlyBaseFee +
    subscription.monthlyFeePerDevice * subscription.currentDevicesAmount;
  const freeCredits = subscription.billingDetails.freeCredits / 100;
  const hasCreditCard = !!subscription.billingDetails.expiryDate;

  let deviceCostPara = <></>;
  if (!freeloader && nextBill > 0) {
    deviceCostPara = (
      <p>
        Next bill: <span className="font-bold">{nextBill}€</span>
      </p>
    );
  }
  let freeCreditsPara = <></>;
  if (!freeloader && freeCredits > 0) {
    freeCreditsPara = (
      <p>
        <span className="font-bold">{freeCredits}€</span> free credits remaining
      </p>
    );
  }

  let ccReminderPara = <></>;
  if (!freeloader && nextBill > freeCredits && !hasCreditCard) {
    ccReminderPara = <p>Please add a credit card for continued use!</p>;
  }

  return (
    <>
      {deviceCostPara}
      {freeCreditsPara}
      {ccReminderPara}
    </>
  );
};
export default SubscriptionSummary;
