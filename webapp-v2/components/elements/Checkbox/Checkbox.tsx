import classNames from 'classnames';
import React, { ForwardRefRenderFunction } from 'react';

interface CheckboxProps extends React.HTMLProps<HTMLInputElement> {
  label?: string;
  onChangeValue?: (s: boolean) => void;
  wrapperClassName?: string;
  errors?: string[];
}

const Checkbox: ForwardRefRenderFunction<HTMLInputElement, CheckboxProps> = (
  { label, onChange, onChangeValue, className, wrapperClassName, errors, disabled, ...rest },
  ref,
) => {
  const inputName = label;
  const hasErrors = errors && errors.filter((err) => err).length > 0;
  return (
    <div className={wrapperClassName}>
      <div className="flex space-x-2">
        <input
          ref={ref}
          type="checkbox"
          name={inputName}
          id={inputName}
          disabled={disabled}
          className={classNames(className, ' inline-block sm:text-sm', {
            'border-red-500 checked:border-red-500 hover:border-red-500': hasErrors,
            'border-gray-800': !hasErrors,
          })}
          onChange={(e) => {
            if (onChange) {
              onChange(e);
            }
            if (onChangeValue) {
              onChangeValue(e.target.checked);
            }
          }}
          {...rest}
        />
        {label && (
          <label
            htmlFor={inputName}
            className={classNames(['block text-sm font-medium text-gray-700'])}>
            {label}
          </label>
        )}
      </div>

      {errors && (
        <span className="block text-sm font-medium text-red-700">{errors.join(', ')}</span>
      )}
    </div>
  );
};
export default React.forwardRef(Checkbox);
