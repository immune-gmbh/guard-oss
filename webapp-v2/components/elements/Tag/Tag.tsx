import { QuestionMarkCircleIcon, ShieldCheckIcon, ExclamationIcon } from '@heroicons/react/solid';
import classNames from 'classnames';

interface ITagProps extends React.HTMLProps<HTMLDivElement> {
  status: 'INFO-GHOST' | 'INFO-BOLD' | 'CRITICAL';
  text: string;
}

const Tag: React.FC<ITagProps> = ({ status, text, className, ...rest }) => {
  const renderIcon = (): JSX.Element => {
    if (text === 'Unresponsive') {
      return <QuestionMarkCircleIcon className="h-3.5 text-yellow-900" />;
    } else if (text === 'Trusted') {
      return <ShieldCheckIcon className="h-3.5 text-green-notification" />;
    } else if (text === 'Untrusted') {
      return <ExclamationIcon className="h-3.5 text-red-notification" />;
    }

    if (status === 'CRITICAL') {
      return <ExclamationIcon className="h-4 min-w-[1rem] text-red-notification" />;
    }

    return null;
  };

  return (
    <div
      className={classNames(
        'inline-flex gap-2.5 rounded-full py-1 w-32 px-2.5 items-center whitespace-nowrap',
        {
          [`bg-red-critical`]: status === 'CRITICAL',
          'border border-gray-400': status === 'INFO-GHOST',
          'hover:bg-gray-200': status === 'INFO-GHOST' && rest.onClick,
          'bg-gray-300': status === 'INFO-BOLD',
          'hover:bg-gray-400': status === 'INFO-BOLD' && rest.onClick,
          'cursor-pointer': rest.onClick,
          'pr-3': !!renderIcon(),
        },
        className,
      )}
      title={text}
      {...rest}>
      {renderIcon()}
      <span
        className={classNames('text-sm overflow-hidden text-ellipsis', {
          'text-red-notification': text === 'Untrusted' || status === 'CRITICAL',
          'text-yellow-900': text === 'Unresponsive',
          'text-green-notification': text === 'Trusted',
          'font-bold': status === 'INFO-BOLD',
        })}>
        {text}
      </span>
    </div>
  );
};

export default Tag;
