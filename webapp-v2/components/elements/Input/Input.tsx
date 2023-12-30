import classNames from 'classnames';
import { cloneElement, forwardRef, ReactElement } from 'react';

interface InputProps extends React.HTMLProps<HTMLInputElement> {
  label?: string;
  qaLabel?: string;
  onChangeValue?: (s: string) => void;
  icon?: ReactElement;
  wrapperClassName?: string;
  errors?: string[];
  theme?: 'light' | 'dark';
}

const Input: React.ForwardRefRenderFunction<HTMLInputElement, InputProps> = (
  {
    label,
    qaLabel,
    onChange,
    onChangeValue,
    icon,
    className,
    wrapperClassName,
    errors,
    disabled,
    theme = 'dark',
    ...rest
  },
  ref,
) => {
  const inputName = label;
  const hasErrors = errors && errors.filter((err) => err).length > 0;
  return (
    <div
      className={classNames('space-y-2', {
        [wrapperClassName]: !!wrapperClassName,
      })}>
      {label && (
        <label
          htmlFor={inputName}
          className={classNames([
            'block text-sm font-medium',
            {
              'text-gray-700': theme === 'dark',
              'text-white': theme === 'light',
              'text-red-700': hasErrors && theme === 'dark',
              'text-red-100': hasErrors && theme === 'light',
            },
          ])}>
          {label}
          {rest.required ? '*' : ''}
        </label>
      )}
      <div className="mt-1 relative rounded-md">
        {icon && (
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            {cloneElement(icon, { className: ' h-4 text-gray-600', 'aria-hidden': 'true' })}
          </div>
        )}
        <input
          ref={ref}
          type="text"
          name={inputName}
          id={inputName}
          disabled={disabled}
          data-qa={qaLabel}
          className={classNames(
            'p-4 border focus:ring-indigo-500 focus:border-indigo-500 block w-full text-sm sm:text-xl rounded text-primary',
            {
              [className]: !!className,
              'pl-8': icon,
              'border-red-500': hasErrors,
              'border-gray-400': !hasErrors,
              'cursor-not-allowed bg-gray-100 text-gray-500': disabled || rest.readOnly,
            },
          )}
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
      </div>
      {errors && (
        <span className="block text-sm font-medium text-red-700">{errors.join(', ')}</span>
      )}
    </div>
  );
};
export default forwardRef(Input);
