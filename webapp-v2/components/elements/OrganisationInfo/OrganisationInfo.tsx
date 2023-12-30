import Headline from 'components/elements/Headlines/Headline';
import { useOrganisation } from 'hooks/organisations';

interface IOrganisationInfoProps {
  organisationId: string;
}

const OrganisationInfo: React.FC<IOrganisationInfoProps> = ({ organisationId }) => {
  const { data: organisation } = useOrganisation({ id: organisationId });

  return (
    <div className="space-y-4">
      <Headline size={4}>Organisation Contact</Headline>
      {organisation && organisation.address && (
        <p className="space-y-2 flex flex-col">
          <span key="invoiceName">{organisation['invoiceName']}</span>
          {['streetAndNumber', 'postalCode', 'city', 'country'].map((attr) => (
            <span key={attr}>{organisation.address[attr]}</span>
          ))}
          <span key="vatNumber">{organisation['vatNumber']}</span>
        </p>
      )}
    </div>
  );
};
export default OrganisationInfo;
