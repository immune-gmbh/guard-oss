import classNames from 'classnames';
import _ from 'lodash';
import React, { ForwardRefRenderFunction } from 'react';

interface RadioProps extends React.HTMLProps<HTMLInputElement> {
  label?: string;
  onChangeValue?: (s: string) => void;
  wrapperClassName?: string;
  errors?: string[];
}

const Radio: ForwardRefRenderFunction<HTMLInputElement, RadioProps> = (
  {
    label,
    onChange,
    onChangeValue,
    className,
    wrapperClassName,
    errors,
    disabled,
    name,
    id,
    ...rest
  },
  ref,
) => {
  const hasErrors = errors && errors.filter((err) => err).length > 0;
  id = id || _.uniqueId('radio-');
  return (
    <div className={wrapperClassName}>
      <div className="flex space-x-2">
        <input
          ref={ref}
          type="radio"
          name={name}
          id={id}
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
              onChangeValue(e.target.value);
            }
          }}
          {...rest}
        />
        {label && (
          <label htmlFor={id} className={classNames(['block text-sm font-medium text-gray-700'])}>
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
export default React.forwardRef(Radio);
