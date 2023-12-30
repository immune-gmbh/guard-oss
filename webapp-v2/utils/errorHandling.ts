import NextJsRoutes from 'generated/NextJsRoutes';
import { useRouter } from 'next/router';
import { toast } from 'react-toastify';
import { UrlObject } from 'url';

export function handleLoadError(
  message: string,
  full = false,
  path: string | UrlObject = NextJsRoutes.dashboardIndexPath,
): null {
  const { push } = useRouter();
  push(path);

  const msg = full ? message : `Could not load ${message}.`;

  toast.error(msg, {
    toastId: message,
  });
  return null;
}
