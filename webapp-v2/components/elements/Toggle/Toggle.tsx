import { Switch } from '@headlessui/react';
import classNames from 'classnames';

export interface IToggleProps {
  checked: boolean;
  disabled?: boolean;
  label?: string;
  onChange?: (value: boolean) => void;
  qaLabel?: string;
  labelColor?: string;
}
const Toggle: React.FC<IToggleProps> = ({
  labelColor,
  checked,
  disabled,
  label,
  qaLabel,
  onChange,
}) => {
  if (!label) {
    label = checked ? 'Enabled' : 'Disabled';
  }
  return (
    <Switch.Group as="div" className="flex items-center group">
      <Switch
        checked={checked}
        disabled={disabled}
        onChange={(value) => !disabled && onChange && onChange(value)}
        data-qa={qaLabel}
        className={classNames(
          checked
            ? 'bg-green-notification group-hover:bg-green-500'
            : 'bg-gray-200 group-hover:bg-gray-300',
          'relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500',
          { 'cursor-not-allowed': disabled },
        )}>
        <span
          aria-hidden="true"
          className={classNames(
            checked ? 'translate-x-5' : 'translate-x-0',
            'pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200',
          )}
        />
      </Switch>
      <Switch.Label as="span" className="ml-3">
        <span
          className={classNames('text-sm font-medium cursor-pointer', {
            color: labelColor ? labelColor : 'text-gray-900',
            'cursor-not-allowed': disabled,
          })}>
          {label}{' '}
        </span>
      </Switch.Label>
    </Switch.Group>
  );
};
export default Toggle;
