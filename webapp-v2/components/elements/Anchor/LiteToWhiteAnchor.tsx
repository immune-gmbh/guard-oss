import classNames from 'classnames';
import React, { forwardRef, Ref } from 'react';

const LiteToWhiteAnchor = forwardRef(
  (
    { children, onClick, href, className = '' }: React.HTMLProps<HTMLAnchorElement>,
    ref: Ref<HTMLAnchorElement>,
  ): JSX.Element => {
    return (
      <a
        ref={ref}
        href={href}
        onClick={onClick}
        className={classNames('text-lg transition-colors hover:text-white', className)}>
        {children}
      </a>
    );
  },
);
LiteToWhiteAnchor.displayName = 'LiteWhiteAnchor';

export default LiteToWhiteAnchor;
