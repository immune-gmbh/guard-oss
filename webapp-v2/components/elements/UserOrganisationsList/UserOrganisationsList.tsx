import { ChevronRightIcon, PencilIcon } from '@heroicons/react/solid';
import Link from 'components/elements/Button/Link';
import DeleteOrganisationModal from 'components/elements/DomainModals/DeleteOrganisationModal';
import Headline from 'components/elements/Headlines/Headline';
import Tag from 'components/elements/Tag/Tag';
import { lightFormat, parseISO } from 'date-fns';
import NextJsRoutes from 'generated/NextJsRoutes';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useDeleteOrganisation } from 'hooks/organisations';
import { useSession } from 'hooks/useSession';
import _ from 'lodash';
import NextLink from 'next/link';
import React, { ReactNode } from 'react';
import { useSWRConfig } from 'swr';
import { SerializedOrganisation } from 'utils/types';

const UserOrganisationsList: React.FC = () => {
  const {
    session: { memberships },
  } = useSession();
  const deleteOrganisation = useDeleteOrganisation();
  const { mutate } = useSWRConfig();

  const paymentStatus = (organisation: SerializedOrganisation): ReactNode => {
    if (
      !organisation.subscription ||
      !organisation.subscription?.periodEnd ||
      organisation.subscription?.actionRequired
    ) {
      return (
        <Tag
          text="Action required on subscription"
          status="CRITICAL"
          className="w-auto gap-1 max-w-full"
        />
      );
    }
    return (
      <Tag
        text={`Next Payment: ${lightFormat(
          parseISO(organisation.subscription?.periodEnd),
          'yyyy-MM-dd',
        )}`}
        status="INFO-GHOST"
      />
    );
  };

  return (
    <div className="space-y-6">
      <Headline size={4}>Organisations</Headline>
      <table className="w-full table-fixed">
        <tbody>
          {memberships &&
            memberships.map((membership) => (
              <tr key={membership.id} className="bg-gray-100 border-b">
                <td className="p-4">{membership?.organisation?.name}</td>
                <td className="p-4 w-[15%]">{_.capitalize(membership.role)}</td>
                <td className="p-4 w-[15%]">{membership?.organisation?.memberCount} Members</td>
                <td className="p-4 w-[35%]">{paymentStatus(membership.organisation)}</td>
                <td className="p-4">
                  <div className="flex h-full items-center space-x-2 float-right">
                    {membership.organisation.canDelete && (
                      <DeleteOrganisationModal
                        organisation={{
                          name: membership.organisation?.name,
                          id: membership.organisation?.id,
                        }}
                        onDelete={({ id }) =>
                          deleteOrganisation
                            .mutate({ id })
                            .then(() => mutate(JsRoutesRails.v2_memberships_path()))
                            .then(() => mutate(JsRoutesRails.v2_session_path()))
                        }
                      />
                    )}
                    {membership.organisation.canEdit && (
                      <NextLink
                        passHref
                        href={{
                          pathname: NextJsRoutes.dashboardOrganisationsIdPath,
                          query: { id: membership.organisation?.id, edit: 1 },
                        }}>
                        <a>
                          <PencilIcon className="h-5 cursor-pointer hover:text-gray-600 text-gray-400" />
                        </a>
                      </NextLink>
                    )}
                    <NextLink
                      passHref
                      href={{
                        pathname: NextJsRoutes.dashboardOrganisationsIdPath,
                        query: { id: membership.organisation?.id },
                      }}>
                      <a>
                        <ChevronRightIcon className="h-5 cursor-pointer hover:text-gray-600 text-gray-400" />
                      </a>
                    </NextLink>
                  </div>
                </td>
              </tr>
            ))}
        </tbody>
      </table>
      <div className="flex">
        <Link theme="MAIN" href={NextJsRoutes.dashboardOrganisationsNewPath}>
          New organisation
        </Link>
      </div>
    </div>
  );
};
export default UserOrganisationsList;
