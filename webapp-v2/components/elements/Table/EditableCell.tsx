import { useEffect, useState } from 'react';

interface IEditableCellProps {
  value;
  row;
  column;
  editCell;
}
export const EditableCell: React.FC<IEditableCellProps> = ({
  value: initialValue,
  row,
  column: { id, options },
  editCell,
}) => {
  // We need to keep and update the state of the cell normally
  const [value, setValue] = useState(initialValue);

  // If the initialValue is changed external, sync it up with our state
  useEffect(() => {
    setValue(initialValue);
  }, [options, initialValue]);

  useEffect(() => {
    if (row.original.isDiscarded) {
      setValue(initialValue);
    }
  }, [row.original.isDiscarded, row.original.isEditing]);

  if (!row.original.isEditing) {
    return <span className="block p-2">{options ? options[initialValue] : value}</span>;
  }

  if (options) {
    return (
      <select
        name={id}
        onChange={(e) => setValue(e.target.value)}
        value={value}
        onBlur={() => editCell(row.index, id, value)}
        className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-b border-gray-700 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
        {Object.entries(options).map(([key, value]) => (
          <option key={key} value={key}>
            {value}
          </option>
        ))}
      </select>
    );
  }

  return (
    <input
      name={id}
      className="bg-white p-2 w-full border-b px-2 border-gray-700"
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onBlur={() => editCell(row.index, id, value)}
    />
  );
};
