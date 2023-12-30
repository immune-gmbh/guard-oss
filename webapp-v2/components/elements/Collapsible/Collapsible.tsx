import { ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import { useState } from 'react';

interface ICollapsible extends React.HTMLAttributes<HTMLDivElement> {
  collapsed?: boolean;
  enabled?: boolean;
  indent?: boolean;
  pre?: React.ReactElement;
}

export function Collapsible({
  collapsed = true,
  children,
  enabled = true,
  indent = true,
  pre = null,
}: ICollapsible): JSX.Element {
  const [isCollapsed, setCollapsed] = useState(collapsed);

  const CollapseIcon = isCollapsed ? ChevronRightIcon : ChevronDownIcon;

  return (
    <div className="h-auto" role="switch">
      <div
        data-qa={isCollapsed ? 'collapsed' : 'open'}
        className={classNames('flex items-start justify-between', {
          'cursor-pointer': enabled,
        })}
        onClick={() => enabled && setCollapsed(!isCollapsed)}>
        {pre}
        {enabled && <CollapseIcon className="h-6" />}
        {!enabled && (
          <div className={classNames({ 'space-x-2 select-text': indent })}>{children} </div>
        )}
      </div>
      <div
        className={classNames({
          'mb-4': !isCollapsed,
          'h-0 overflow-hidden': isCollapsed,
        })}>
        <div className={classNames({ 'space-x-2 select-text': indent })}>{children}</div>
      </div>
    </div>
  );
}
