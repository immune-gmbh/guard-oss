import {
  CheckIcon,
  PencilIcon,
  UserGroupIcon,
  SearchIcon,
  RefreshIcon,
  XIcon,
} from '@heroicons/react/solid';
import DeleteOrganisationModal from 'components/elements/DomainModals/DeleteOrganisationModal';
import Table, { TextCell } from 'components/elements/Table/Table';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useDeleteOrganisation, useUpdateOrganisation } from 'hooks/organisations';
import router from 'next/router';
import { FC, useState, useMemo, ReactNode, useEffect, useCallback } from 'react';
import { toast } from 'react-toastify';
import { SerializedOrganisation } from 'utils/types';

interface IOrganisationsTableProps {
  organisations: SerializedOrganisation[];
  searchText: string;
}

const OrganisationsTable: FC<IOrganisationsTableProps> = ({ organisations, searchText }) => {
  const [data, setData] = useState(organisations || []);
  useEffect(() => {
    setData(organisations);
  }, [organisations]);
  const deleteOrganisation = useDeleteOrganisation();
  const updateOrganisation = useUpdateOrganisation();

  const handleDeleteOrganisation = useCallback(
    (rowData): void => {
      deleteOrganisation.mutate({ id: rowData.id }).then((response?: { code: number }) => {
        if (response?.code && response?.code === 200) {
          toast.success('Organisation was successfully deleted.');
          setSkipPageReset(true);
          setData((data) => data.filter((row) => row.id != rowData.id));
        } else {
          toast.error('Something went wrong');
        }
      });
    },
    [deleteOrganisation],
  );
  const handleClickEditRow = useCallback((rowIndex: number): void => {
    setSkipPageReset(true);
    setData((prev) =>
      prev.map((prevRow, index) => ({ ...prevRow, isEditing: rowIndex === index })),
    );
  }, []);
  const handleDiscardEditRow = useCallback(() => {
    setSkipPageReset(true);
    setData((prev) =>
      prev.map((prevRow) => ({ ...prevRow, edits: {}, isEditing: false, isDiscarded: true })),
    );
  }, []);
  const handleAcceptEditRow = useCallback(
    async (rowIndex, rowData) => {
      // Return if no changes were made
      setSkipPageReset(true);
      if (!rowData.edits || Object.keys(rowData.edits).length == 0) {
        setData((prev) =>
          prev.map((prevRow) => ({ ...prevRow, edits: {}, isEditing: false, isDiscarded: false })),
        );
        return;
      }

      const previousData = data;
      // Optimistic Update
      setData((prev) =>
        prev.map((prevRow) =>
          rowData.id == prevRow.id
            ? {
                ...prevRow,
                ...rowData.edits,
                isEditing: false,
                isLoading: true,
                isDiscarded: false,
              }
            : { ...prevRow, isEditing: false, isLoading: false, isDiscarded: false },
        ),
      );

      const updateResponse = (await updateOrganisation.mutate({
        ...rowData.edits,
        id: rowData.id,
      })) as any;

      if (updateResponse.errors) {
        toast.error('Something went wrong');
        setData(previousData);
      } else {
        toast.success('Organisation was successfully updated.');
        setSkipPageReset(true);
        setData((prev) =>
          prev.map((prevRow) =>
            rowData.id == prevRow.id
              ? {
                  ...prevRow,
                  ...updateResponse,
                  isEditing: false,
                  isLoading: false,
                  ...rowData.edits,
                }
              : { ...prevRow, isEditing: false, isLoading: false },
          ),
        );
      }
    },
    [updateOrganisation],
  );

  const ActionsComponent = ({ row }): ReactNode => (
    // Use Cell to render an expander for each row.
    // We can use the getToggleRowExpandedProps prop-getter
    // to build the expander.
    <span className="flex space-x-4">
      {row.original.isLoading && <RefreshIcon className="w-6 animate-spin text-gray-400" />}
      {!row.original.isLoading &&
        ((row.original.isEditing && (
          <>
            <CheckIcon
              className="w-6 cursor-pointer hover:text-green-800 text-green-600"
              data-qa="org-accept"
              onClick={() => {
                handleAcceptEditRow(row.index, row.original);
              }}
            />
            <XIcon
              className="w-6 cursor-pointer hover:text-red-800 text-red-600"
              data-qa="org-abort"
              onClick={() => handleDiscardEditRow()}
            />
          </>
        )) || (
          <>
            <SearchIcon
              className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
              data-qa="org-search"
              onClick={() =>
                router.push({
                  pathname: NextJsRoutes.adminOrganisationsIdIndexPath,
                  query: { id: row.original.id },
                })
              }
            />
            <UserGroupIcon
              className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
              data-qa="org-users"
              onClick={() =>
                router.push({
                  pathname: NextJsRoutes.adminOrganisationsIdUsersPath,
                  query: { id: row.original.id },
                })
              }
            />
            <DeleteOrganisationModal
              organisation={row.original}
              onDelete={handleDeleteOrganisation}
            />
            <PencilIcon
              className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
              data-qa="org-edit"
              onClick={() => handleClickEditRow(row.index)}
            />
          </>
        ))}
    </span>
  );
  const columns = useMemo(() => {
    return [
      {
        Header: 'Name',
        accessor: 'name',
      },
      {
        Header: 'Status',
        accessor: 'status',
        options: {
          created: 'created',
          active: 'active',
          suspended: 'suspended',
        },
      },
      {
        Header: 'Freeloader',
        accessor: 'freeloader',
        options: {
          true: 'Yes',
          false: 'No',
        },
      },
      {
        Header: 'Members',
        accessor: 'memberships.length',
        Cell: TextCell,
      },
      {
        Header: 'Devices',
        accessor: 'subscription.currentDevicesAmount',
        Cell: TextCell,
      },
      {
        // Make an expander cell
        Header: () => null, // No header
        id: 'actions', // It needs an ID
        disableSortBy: true,
        Cell: ActionsComponent,
      },
    ];
    // Clicking an action in ActionsComponent does not work if we add it to the depencylist
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const [skipPageReset, setSkipPageReset] = useState(false);

  // We need to keep the table from resetting the pageIndex when we
  // Update data. So we can keep track of that flag with a ref.

  // When our cell renderer calls editCell, we'll use
  // the rowIndex, columnId and new value to update the
  // original data
  const editCell = (rowIndex: number, columnId: number, value): void => {
    // We also turn on the flag to not reset the page
    setSkipPageReset(true);
    setData((old) =>
      old.map((row, index) => {
        if (index === rowIndex) {
          return {
            ...old[rowIndex],
            edits: {
              ...old[rowIndex]['edits'],
              [columnId]: value,
            },
          };
        }
        return row;
      }),
    );
  };

  // After data chagnes, we turn the flag back off
  // so that if data actually changes when we're not
  // editing it, the page is reset
  useEffect(() => {
    setSkipPageReset(false);
  }, [data]);

  return (
    <Table
      columns={columns}
      data={data}
      editCell={editCell}
      skipPageReset={skipPageReset}
      searchText={searchText}
    />
  );
};
export default OrganisationsTable;
