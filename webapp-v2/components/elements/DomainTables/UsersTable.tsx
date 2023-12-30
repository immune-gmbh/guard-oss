import { CheckIcon, PencilIcon, RefreshIcon, XIcon } from '@heroicons/react/solid';
import DeleteUserModal from 'components/elements/DomainModals/DeleteUserModal';
import Table from 'components/elements/Table/Table';
import { useSession } from 'hooks/useSession';
import { useDeleteUser, useUpdateUser } from 'hooks/users';
import { FC, useState, useMemo, ReactNode, useEffect, useCallback } from 'react';
import { toast } from 'react-toastify';
import { SerializedUser } from 'utils/types';

interface IUsersTableProps {
  users: SerializedUser[];
  searchText: string;
  roleFilter: string;
}

const filterUsers = (users: SerializedUser[], roleFilter: string): SerializedUser[] =>
  users.filter((user) => roleFilter == 'all' || user.role == roleFilter);

const UsersTable: FC<IUsersTableProps> = ({ users, searchText, roleFilter }) => {
  const [data, setData] = useState(filterUsers(users, roleFilter) || []);
  useEffect(() => {
    setData(filterUsers(users, roleFilter));
  }, [users, roleFilter]);

  const {
    session: {
      user: { id: currentUserId },
    },
  } = useSession();

  const deleteUser = useDeleteUser();
  const updateUser = useUpdateUser();

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
    [updateUser],
  );

  const ActionsComponent = ({ row }): ReactNode => (
    // Use Cell to render an expander for each row.
    // the user information is attached to row.original
    // We can use the getToggleRowExpandedProps prop-getter
    // to build the expander.
    <span className="flex space-x-4">
      {row.original.isLoading && <RefreshIcon className="w-6 animate-spin text-gray-400" />}
      {!row.original.isLoading &&
        ((row.original.isEditing && (
          <>
            <CheckIcon
              className="w-6 cursor-pointer hover:text-green-800 text-green-600"
              onClick={() => {
                handleAcceptEditRow(row.index, row.original);
              }}
            />
            <XIcon
              className="w-6 cursor-pointer hover:text-red-800 text-red-600"
              onClick={() => handleDiscardEditRow()}
            />
          </>
        )) || (
          <>
            {row.original.id !== currentUserId && (
              <DeleteUserModal user={row.original} onDelete={handleDeleteUser} />
            )}
            <PencilIcon
              className="w-6 cursor-pointer hover:text-gray-600 text-gray-400"
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
        Header: 'Email',
        accessor: 'email',
      },
      {
        Header: 'Organisation(s)',
        accessor: (user: SerializedUser) => {
          return user.organisations.map((orga) => orga.name).join(', ');
        },
      },
      {
        Header: 'Role',
        accessor: 'role',
        options: { user: 'User', admin: 'Admin' },
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
export default UsersTable;
