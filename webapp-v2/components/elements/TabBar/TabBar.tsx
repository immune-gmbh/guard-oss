import { Children } from 'react';

export default function TabBar({ children }: React.HTMLProps<HTMLDivElement>): JSX.Element {
  return (
    <div className="grid items-center justify-around bg-purple-300 bg-opacity-50 h-12">
      {children}

      <style jsx>{`
        div {
          grid-template-columns: repeat(${Children.count(children)}, minmax(0, 1fr));
        }
      `}</style>
    </div>
  );
}
