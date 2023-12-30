import classNames from 'classnames';
import { createElement } from 'react';

type HeadlineSize = 1 | 2 | 3 | 4 | 5;

interface IHeadline extends React.HTMLAttributes<HTMLHeadingElement> {
  size?: HeadlineSize;
  as?: 'h1' | 'h2' | 'h3' | 'h4';
  bold?: boolean;
}

export default function Headline({
  children,
  className,
  size,
  as = 'h2',
  bold = true,
}: IHeadline): JSX.Element {
  let useSize = size;
  if (!size) {
    useSize = parseInt(as.replace('h', ''), 10) as HeadlineSize;
  }

  const headlineClasses = classNames({
    'font-bold': (useSize === 1 || useSize === 3) && bold,
    'text-5xl leading-tight': useSize === 1 || useSize === 2,
    'text-4xl leading-snug': useSize === 3,
    'text-3xl': useSize === 4,
    'text-2xl': useSize === 5,
    [className]: !!className,
  });

  return createElement(as, { className: headlineClasses }, children);
}
