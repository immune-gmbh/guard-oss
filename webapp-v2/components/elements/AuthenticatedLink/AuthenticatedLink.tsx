import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import { useRef, useState } from 'react';
import { fetchAuthToken } from 'utils/session';

interface IAuthenticatedLink {
  url: string;
  filename: string;
  qaLabel?: string;
}

const AuthenticatedLink: React.FC<IAuthenticatedLink> = ({ url, filename, qaLabel, children }) => {
  const link = useRef<HTMLAnchorElement>();

  const [isLoading, setIsLoading] = useState(false);

  const handleAction = async (): Promise<void> => {
    if (!link?.current || link.current.href) {
      return;
    }

    setIsLoading(true);

    const result = await fetch(url, {
      headers: { Authorization: `Bearer ${fetchAuthToken()}` },
    });

    const blob = await result.blob();
    const href = window.URL.createObjectURL(blob);

    link.current.download = filename;
    link.current.href = href;

    link.current.click();
    setIsLoading(false);
  };

  return (
    <>
      <a
        role="button"
        ref={link}
        data-qa={qaLabel}
        onClick={handleAction}
        className="font-bold underline cursor-pointer">
        {isLoading && <LoadingSpinner />}
        {!isLoading && children}
      </a>
    </>
  );
};
export default AuthenticatedLink;
