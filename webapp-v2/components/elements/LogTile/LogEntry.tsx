import { formatDistance } from 'date-fns';
import NextJsRoutes from 'generated/NextJsRoutes';
import Link from 'next/link';
import { ApiSrv } from 'types/apiSrv';

interface ILogEntryProps {
  change: ApiSrv.Change;
}

const LogEntry: React.FC<ILogEntryProps> = ({ change }: ILogEntryProps) => {
  let message: JSX.Element;
  let actor = '';

  if (
    change.actor !== '(No actor)' &&
    change.actor !== 'agent' &&
    !change.actor.startsWith('tag:immu.ne,2021')
  ) {
    actor = ` by ${change.actor}`;
  }

  switch (change.type) {
    case 'enroll':
      message = (
        <>
          New device{' '}
          <Link
            passHref
            href={{
              pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
              query: { did: change.devices?.id },
            }}>
            <a>enrolled{actor}.</a>
          </Link>
        </>
      );
      break;
    case 'resurrect':
      message = (
        <>
          Reverted retire of{' '}
          <Link
            passHref
            href={{
              pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
              query: { did: change.devices?.id },
            }}>
            <a>
              device #{change.devices.id}
              {actor}.
            </a>
          </Link>
        </>
      );
      break;
    case 'rename':
      if (change.devices) {
        message = (
          <>
            Renamed{' '}
            <Link
              passHref
              href={{
                pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
                query: { did: change.devices.id },
              }}>
              <a>
                device #{change.devices.id}
                {actor}.
              </a>
            </Link>
          </>
        );
      }
      break;
    case 'tag':
      message = (
        <>
          Tag{' '}
          <Link
            passHref
            href={{
              pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
              query: { did: change.devices.id },
            }}>
            <a>
              device #{change.devices.id}
              {actor}.
            </a>
          </Link>
        </>
      );
      break;
    case 'retire':
      message = (
        <>
          Retired{' '}
          <Link
            passHref
            href={{
              pathname: NextJsRoutes.dashboardDevicesDidIndexPath,
              query: { did: change.devices.id },
            }}>
            <a>
              device #{change.devices.id}
              {actor}.
            </a>
          </Link>
        </>
      );
      break;
    default:
      return <></>;
  }
  const timeAgo = formatDistance(new Date(), new Date(change.timestamp));

  return (
    <p key={change.id}>
      <span className="font-bold">{timeAgo} ago&nbsp;</span>
      <span>{message}</span>
    </p>
  );
};

export const willBeRendered = (change: ApiSrv.Change): boolean => {
  return (
    {
      enroll: true,
      resurrect: true,
      rename: true,
      tag: true,
      retire: true,
    }[change.type] || false
  );
};

export default LogEntry;
