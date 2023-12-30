import { ReactNode, useLayoutEffect, useState } from 'react';
import { createPortal } from 'react-dom';

const Actionbar: React.FC = ({ children }) => {
  const [root, setRoot] = useState(document.getElementById('actionbar-slot'));

  useLayoutEffect(() => {
    setRoot(document.getElementById('actionbar-slot'));
  }, []);

  if (document.getElementsByClassName('actionbar').length > 1) {
    throw new Error('Multiple Actionbars rendered');
  }

  if (!root) return null;

  const Actionbar: ReactNode = (
    <div className="p-5 flex justify-between actionbar min-h-[84px]">{children}</div>
  );
  return createPortal(Actionbar, root);
};

export default Actionbar;
