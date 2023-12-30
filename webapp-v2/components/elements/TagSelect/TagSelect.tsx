import Tag from 'components/elements/Tag/Tag';
import { cloneElement, FC, ReactElement } from 'react';

interface ITagSelectProps<T extends string> {
  options: Record<T, string>;
  selectedKey: T;
  onSelect?: (key: T) => void;
  icon?: ReactElement;
}

const TagSelect: FC<ITagSelectProps<string>> = ({ options, selectedKey, onSelect, icon }) => {
  return (
    <div className="flex space-x-2 items-center">
      {icon && cloneElement(icon, { className: ' h-4 text-purple-300', 'aria-hidden': 'true' })}
      {Object.entries(options).map(([key, text]) => (
        <Tag
          key={key}
          onClick={() => onSelect && onSelect(key)}
          text={text}
          status={key == selectedKey ? 'INFO-BOLD' : 'INFO-GHOST'}
        />
      ))}
    </div>
  );
};
export default TagSelect;
