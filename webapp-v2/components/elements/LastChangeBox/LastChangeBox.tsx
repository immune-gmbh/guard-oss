import { ClockIcon } from '@heroicons/react/outline';
import { formatDistance } from 'date-fns';
import { ApiSrv } from 'types/apiSrv';

interface ILastChangeBox {
  change: ApiSrv.Change;
}

const changeToText = (change: ApiSrv.Change): { retire: string }[] | string => {
  return (
    {
      retire: 'Retired',
      enroll: 'Enrolled',
      resurrect: 'Undeleted',
      rename: 'Renamed',
      tag: 'Tagged',
      associate: 'Policy changed',
    }[change.type] || 'Changed'
  );
};

export default function LastChangeBox({ change }: ILastChangeBox): JSX.Element {
  if (!change) return null;

  const timeAgo = formatDistance(new Date(), new Date(change.timestamp));

  return (
    <div className="absolute top-8 right-8 flex py-2 px-4 rounded space-x-2 items-center font-bold bg-gray-300">
      <ClockIcon width={20} />
      <span>{`${changeToText(change)} ${timeAgo} ago`}</span>
    </div>
  );
}
