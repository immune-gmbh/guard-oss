import NoInformationCard from 'components/elements/NoInformationCard/NoInformationCard';
import TrustChainBar from 'components/elements/TrustChainBar/TrustChainBar';

export default function DeviceNoAppraisal(): JSX.Element {
  return (
    <>
      <TrustChainBar unknown={true} />
      <div className="mt-16">
        <NoInformationCard />
      </div>
    </>
  );
}
