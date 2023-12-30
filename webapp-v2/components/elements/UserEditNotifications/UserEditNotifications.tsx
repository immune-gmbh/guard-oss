import Headline from 'components/elements/Headlines/Headline';
import Spinner from 'components/elements/Spinner/Spinner';
import Toggle from 'components/elements/Toggle/Toggle';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useMemberships, useUpdateMembership } from 'hooks/memberships';
import { useSWRConfig } from 'swr';
import { SerializedMembership } from 'utils/types';

const UserEditNotifications: React.FC = () => {
  const { data: memberships, isLoading } = useMemberships();
  const updateMembership = useUpdateMembership();
  const { mutate } = useSWRConfig();

  return (
    <div className="space-y-6">
      <Headline size={4}>Organisations</Headline>
      <p>Do you want to receive information via mail? </p>
      {isLoading && <Spinner />}
      <table role="table" className="w-full table-fixed">
        <tbody>
          {memberships &&
            memberships.map((membership) => (
              <tr key={membership.id} className="bg-gray-100 border-b">
                <td className="p-4">{membership.organisation.name}</td>
                <td className="p-4">
                  <Toggle
                    qaLabel="billing"
                    onChange={(value) => {
                      updateMembership
                        .mutate({ id: membership.id, membership: { notifyInvoice: value } })
                        .then((updatedMembership: SerializedMembership) =>
                          mutate(
                            JsRoutesRails.v2_memberships_path(),
                            memberships.map((someMembership) =>
                              someMembership.id == updatedMembership.id
                                ? updatedMembership
                                : someMembership,
                            ),
                            false,
                          ),
                        );
                    }}
                    label="Invoice"
                    checked={membership.notifyInvoice}
                  />
                </td>
                <td className="p-4">
                  <Toggle
                    qaLabel="alerts"
                    onChange={(value) => {
                      updateMembership
                        .mutate({ id: membership.id, membership: { notifyDeviceUpdate: value } })
                        .then((updatedMembership: SerializedMembership) =>
                          mutate(
                            JsRoutesRails.v2_memberships_path(),
                            memberships.map((someMembership) =>
                              someMembership.id == updatedMembership.id
                                ? updatedMembership
                                : someMembership,
                            ),
                            false,
                          ),
                        );
                    }}
                    label="Device Updates"
                    checked={membership.notifyDeviceUpdate}
                  />
                </td>
              </tr>
            ))}
        </tbody>
      </table>
    </div>
  );
};
export default UserEditNotifications;
