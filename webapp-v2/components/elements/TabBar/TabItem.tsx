import classNames from 'classnames';

interface ITabItem extends React.HTMLProps<HTMLDivElement> {
  active?: boolean;
  title: string;
  onClick: () => void;
}

export default function TabItem({ title, active, onClick }: ITabItem): JSX.Element {
  return (
    <>
      <div
        className={classNames(
          'flex h-full items-center justify-center px-3 bg-purple-300 bg-opacity-0 transition-colors hover:cursor-pointer hover:bg-opacity-50',
          {
            'border-b-2 border-primary font-bold': active,
          },
        )}
        onClick={onClick}>
        {title}
      </div>
    </>
  );
}
