import classNames from 'classnames';

export default function SubHeadline({
  children,
  className,
}: React.HTMLAttributes<HTMLSpanElement>): JSX.Element {
  return (
    <span
      className={classNames('block text-sm font-bold uppercase tracking-super mb-2', {
        [className]: !!className,
      })}>
      {children}
    </span>
  );
}
