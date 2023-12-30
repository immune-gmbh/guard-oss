import { useSession } from 'hooks/useSession';
import { useEffect } from 'react';

export default function Logout(): void {
  const { logout } = useSession();

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => logout(), []);

  return null;
}
