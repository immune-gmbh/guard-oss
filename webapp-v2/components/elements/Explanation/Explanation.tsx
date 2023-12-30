import { useState } from 'react';

export function Explanation({ children }: React.HTMLAttributes<HTMLDivElement>): JSX.Element {
  const [show, setShow] = useState(false);

  return (
    <div className="border-t border-purple-500 mt-6 pt-6">
      {show && <div className="mb-6">{children}</div>}
      <button className="block underline text-xl font-bold" onClick={() => setShow(!show)}>
        {show ? 'Hide' : 'Show'} explanation
      </button>
    </div>
  );
}
