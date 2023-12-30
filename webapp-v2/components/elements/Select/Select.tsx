import classNames from 'classnames';
import React, { ForwardRefRenderFunction } from 'react';

interface ISelectProps extends React.HTMLProps<HTMLSelectElement> {
  label?: string;
  qaLabel?: string;
  onChangeValue?: (s: string) => void;
  selectedOption?: string;
  options: Record<string, string>;
  errors?: string[];
  wrapperClassName?: string;
  theme?: 'default' | 'light';
}

// For now, use system select box. Swap implementation and styling if needed
const Select: ForwardRefRenderFunction<HTMLSelectElement, ISelectProps> = (
  {
    label,
    qaLabel,
    options,
    onChange,
    onChangeValue,
    errors,
    disabled,
    readOnly,
    selectedOption,
    wrapperClassName,
    theme = 'default',
    ...rest
  },
  ref,
) => {
  const hasErrors = errors && errors.filter((err) => err).length > 0;

  const selectClasses = {
    default: classNames([
      'block pl-3 pr-10 py-4 text-sm sm:text-xl border rounded focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm',
      {
        'border-red-500': hasErrors,
        'border-gray-800': !hasErrors,
        'cursor-not-allowed bg-gray-100 text-gray-500': disabled || readOnly,
      },
    ]),
    light: classNames([
      'block text-gray-600 text-sm sm:text-xl pl-3 pr-10 py-4 border-none rounded focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-base',
      {
        'border-red-500': hasErrors,
        'border-gray-800': !hasErrors,
        'cursor-not-allowed': disabled,
      },
    ]),
  };

  return (
    <div className={classNames('space-y-2 flex flex-col ', wrapperClassName)}>
      {label && (
        <span
          className={classNames([
            'block text-sm font-medium text-gray-700',
            { 'text-red-700': hasErrors },
          ])}>
          {label}
        </span>
      )}
      <select
        style={{
          boxSizing: 'content-box',
        }}
        data-qa={qaLabel}
        role="listbox"
        disabled={disabled || readOnly}
        className={selectClasses[theme]}
        defaultValue={selectedOption}
        onChange={(e) => {
          if (onChange) {
            onChange(e);
          }
          if (onChangeValue) {
            onChangeValue(e.target.value);
          }
        }}
        ref={ref}
        {...rest}>
        {Object.entries(options).map(([key, value]) => (
          <option key={key} value={key}>
            {value}
          </option>
        ))}
      </select>
      {readOnly && <input type="hidden" value={rest.value} />}
      {errors && (
        <span className="block text-sm font-medium text-red-700">{errors.join(', ')}</span>
      )}
    </div>
  );
};
export default React.forwardRef(Select);
