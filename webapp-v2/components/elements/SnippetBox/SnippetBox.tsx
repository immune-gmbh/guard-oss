import classNames from 'classnames';

interface ISnippetBox extends React.HTMLAttributes<HTMLDivElement> {
  withBg?: boolean;
}

export default function SnippetBox({
  children,
  className,
  withBg = false,
  ...attributes
}: ISnippetBox): JSX.Element {
  return (
    <div
      className={classNames(
        'flex shadow-lg',
        {
          'flex-col': !className?.includes('flex-row'),
          'p-8': !className?.match(/p-[0-9]{1,2}/gm),
        },
        {
          'bg-cell-small-solid bg-no-repeat bg-right-bottom bg-contain': withBg,
          [className]: !!className,
        },
      )}
      {...attributes}>
      {children}
    </div>
  );
}
