import { SortAscendingIcon, SortDescendingIcon } from '@heroicons/react/outline';
import React, { useEffect } from 'react';
import { useAsyncDebounce, useGlobalFilter, usePagination, useSortBy, useTable } from 'react-table';

import { EditableCell } from './EditableCell';
import { Pagination } from './Pagination';

export const TextCell: React.FC<{ value: string }> = ({ value }) => <span>{value}</span>;

// Set our editable cell renderer as the default Cell renderer
const defaultColumn = {
  Cell: EditableCell,
};

interface TableProps {
  columns;
  data;
  editCell;
  skipPageReset: boolean;
  searchText: string;
  paginated?: boolean;
}

const Table: React.FC<TableProps> = ({
  columns,
  data,
  editCell,
  skipPageReset,
  searchText,
  paginated,
}) => {
  // For this example, we're using pagination to illustrate how to stop
  // the current page from resetting when our data changes
  // Otherwise, nothing is different here.
  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    prepareRow,
    page,
    canPreviousPage,
    canNextPage,
    pageOptions,
    pageCount,
    gotoPage,
    nextPage,
    previousPage,
    setPageSize,
    state: { pageIndex, pageSize },
    setGlobalFilter,
  } = useTable(
    {
      columns,
      data,
      defaultColumn,
      // use the skipPageReset option to disable page resetting temporarily
      autoResetPage: !skipPageReset,
      autoResetSortBy: !skipPageReset,
      autoResetFilters: !skipPageReset,
      autoResetGlobalFilter: !skipPageReset,
      // editCell isn't part of the API, but
      // anything we put into these options will
      // automatically be available on the instance.
      // That way we can call this function from our
      // cell renderer!
      editCell,
    },
    useGlobalFilter, // useGlobalFilter!
    useSortBy,
    usePagination,
  );

  // Render the UI for your table
  const asyncSetGlobalFilter = useAsyncDebounce((searchText) => {
    setGlobalFilter(searchText || undefined);
  }, 50);
  useEffect(() => {
    asyncSetGlobalFilter(searchText);
  }, [asyncSetGlobalFilter, searchText]);
  useEffect(() => {
    !paginated && setPageSize(9999);
  }, [setPageSize, paginated]);

  return (
    <>
      <table className="styled-table" {...getTableProps()}>
        <thead>
          {headerGroups.map((headerGroup) => (
            // eslint-disable-next-line react/jsx-key
            <tr {...headerGroup.getHeaderGroupProps()}>
              {headerGroup.headers.map((column) => (
                // eslint-disable-next-line react/jsx-key
                <th {...column.getHeaderProps(column.getSortByToggleProps())}>
                  {!column.disableSortBy && (
                    <span className="inline mr-2">
                      {column.isSorted ? (
                        column.isSortedDesc ? (
                          <SortAscendingIcon className="text-purple-100 w-6 inline" />
                        ) : (
                          <SortDescendingIcon className="text-purple-100 w-6 inline" />
                        )
                      ) : (
                        <SortDescendingIcon className="text-gray-400 w-6 inline" />
                      )}
                    </span>
                  )}
                  {column.render('Header')}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody {...getTableBodyProps()}>
          {page.map((row) => {
            prepareRow(row);
            return (
              // eslint-disable-next-line react/jsx-key
              <tr {...row.getRowProps()}>
                {row.cells.map((cell) => {
                  // eslint-disable-next-line react/jsx-key
                  return <td {...cell.getCellProps()}>{cell.render('Cell')}</td>;
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
      {paginated && (
        <Pagination
          gotoPage={gotoPage}
          canPreviousPage={canPreviousPage}
          previousPage={previousPage}
          nextPage={nextPage}
          canNextPage={canNextPage}
          pageCount={pageCount}
          pageIndex={pageIndex}
          pageOptions={pageOptions}
          pageSize={pageSize}
          setPageSize={setPageSize}
        />
      )}
    </>
  );
};
export default Table;
