import { CheckIcon, PencilIcon, RefreshIcon, XIcon } from '@heroicons/react/solid';
import DeleteUserModal from 'components/elements/DomainModals/DeleteUserModal';
import Table, { TextCell } from 'components/elements/Table/Table';
import { useDeleteMembership, useUpdateMembership } from 'hooks/memberships';
import { useSession } from 'hooks/useSession';
import { FC, ReactNode, useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'react-toastify';
import { SerializedMembership } from 'utils/types';

interface IMembershipsTableProps {
  memberships: SerializedMembership[];
  searchText: string;
  roleFilter: string;
  editable: boolean;
  adminView: boolean;
}

const filterUsers = (
  memberships: SerializedMembership[],
  roleFilter: string,
): SerializedMembership[] =>
  memberships.filter((membership) => roleFilter == 'all' || membership.role == roleFilter);

const MembershipsTable: FC<IMembershipsTableProps> = ({
  memberships,
  searchText,
  roleFilter,
  editable,
  adminView,
}) => {
  const [data, setData] = useState(filterUsers(memberships, roleFilter) || []);
  useEffect(() => {
    setData(filterUsers(memberships, roleFilter));
  }, [memberships, roleFilter]);

  const deleteUser = useDeleteMembership();
  const updateUser = useUpdateMembership();
  const {
    session: {
      user: { id: currentUserId },
    },
  } = useSession();

  const handleDeleteUser = useCallback(
    (rowData): void => {
      deleteUser.mutate({ id: rowData.id }).then(() => {
        toast.success('User was successfully deleted.');
        setSkipPageReset(true);
        setData((data) => data.filter((row) => row.id != rowData.id));
      });
    },
    [deleteUser],
  );
  const handleClickEditRow = useCallback((rowIndex: number): void => {
    setSkipPageReset(true);
    setData((prev) =>
      prev.map((prevRow, index) => ({ ...prevRow, isEditing: rowIndex === index })),
    );
  }, []);
  const handleDiscardEditRow = useCallback(() => {
    setSkipPageReset(true);
    setData((prev) => prev.map((prevRow) => ({ ...prevRow, edits: {}, isEditing: false })));
  }, []);
  const handleAcceptEditRow = useCallback(
    (rowIndex, rowData): void => {
      // Return if no changes were made
      setSkipPageReset(true);
      if (!rowData.edits || Object.keys(rowData.edits).length == 0) {
        setData((prev) => prev.map((prevRow) => ({ ...prevRow, edits: {}, isEditing: false })));
        return;
      }
      // Optimistic Update
      setData((prev) =>
        prev.map((prevRow) =>
          rowData.id == prevRow.id
            ? { ...prevRow, ...rowData.edits, isEditing: false, isLoading: true }
            : { ...prevRow, isEditing: false, isLoading: false },
        ),
      );

      // filter changes to admins
      if (!adminView && rowData.role === 'admin' && rowData.edits.role) {
        delete rowData.edits.role;
      }

      updateUser.mutate({ ...rowData.edits, id: rowData.id }).then((...res) => {
        toast.success('User was successfully updated.');
        setSkipPageReset(true);
        setData((prev) =>
          prev.map((prevRow) =>
            rowData.id == prevRow.id
              ? { ...prevRow, ...res, isEditing: false, isLoading: false, ...rowData.edits }
              : { ...prevRow, isEditing: false, isLoading: false },
          ),
        );
      });
    },
    [updateUser, adminView],
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
              data-qa="edit-accept"
              onClick={() => {
                handleAcceptEditRow(row.index, row.original);
              }}
            />
            <XIcon
              className="w-6 cursor-pointer hover:text-red-800 text-red-600"
              data-qa="edit-cancel"
              onClick={() => handleDiscardEditRow()}
            />
          </>
        )) || (
          <>
            {row.original.canDelete && row.original.user.id !== currentUserId && (
              <DeleteUserModal user={row.original} onDelete={handleDeleteUser} />
            )}
            {editable && (
              <PencilIcon
                className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
                data-qa="user-edit"
                onClick={() => handleClickEditRow(row.index)}
              />
            )}
          </>
        ))}
    </span>
  );
  const columns = useMemo(() => {
    const columns = [
      {
        Header: 'Name',
        accessor: 'user.name',
        Cell: TextCell,
      },
      {
        Header: 'Email',
        accessor: 'user.email',
        Cell: TextCell,
      },
      {
        // hide the admin role if we're not an admin
        Header: 'Role',
        accessor: (row: SerializedMembership) =>
          !adminView && row.role === 'admin' ? 'owner' : row.role,
        id: 'role',
        options: Object.assign(
          { user: 'User', owner: 'Owner' },
          adminView ? { admin: 'Admin' } : {},
        ),
      },
      {
        // Make an expander cell
        Header: () => null, // No header
        id: 'actions', // It needs an ID
        disableSortBy: true,
        Cell: ActionsComponent,
      },
    ];
    return columns;
    // Clicking an action in ActionsComponent does not work if we add it to the depencylist
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [adminView]);
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
export default MembershipsTable;
