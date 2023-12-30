import { ChevronDownIcon, ChevronRightIcon, EyeIcon } from '@heroicons/react/solid';
import ImmuneLink from 'components/elements/Button/Link';
import { useIssueFromId } from 'components/elements/Issues/Index';
import { ISSUE_STATUS_STEPS } from 'components/elements/TrustChainBar/TrustChainIcon';
import NextJsRoutes from 'generated/NextJsRoutes';
import { IIncidentsRanking, IIssueRanking } from 'hooks/dashboard';
import useTranslation from 'next-translate/useTranslation';
import { useState } from 'react';
import { rfcToStringFormat, timeAgoFormat } from 'utils/datetime';
import { aspectByIssueId } from 'utils/issues';

export const IssueTableHeader = (): JSX.Element => {
  const { t } = useTranslation();

  return (
    <div className="grid gap-2 grid-cols-incidents-table mt-10 px-4" role="table">
      <b className="mb-4">{t('dashboard:issuesTable.name')}</b>
      <b className="text-right">{t('dashboard:issuesTable.component')}</b>
      <b className="text-right">{t('dashboard:issuesTable.devices')}</b>
      <b></b>
    </div>
  );
};

export const IssueTableRow = ({ issue }: { issue: IIssueRanking }): JSX.Element => {
  const { t } = useTranslation();

  const [toggle, setToggle] = useState(false);

  const { slug, description } = useIssueFromId(issue.issueId);
  const aspectKey = aspectByIssueId(issue.issueId);

  const toggleDescription = (): void => setToggle(!toggle);

  const timestamp = (issue as IIncidentsRanking).timestamp;

  return (
    <div
      className="grid gap-2 gap-y-5 grid-cols-incidents-table even:bg-gray-100 p-4"
      key={issue.issueId}>
      <div>
        <b>{slug}</b>
        {!!timestamp && (
          <small title={rfcToStringFormat(timestamp)} className="text-purple-100 block">
            {timeAgoFormat(timestamp, t)}
          </small>
        )}
      </div>
      <div>
        <div className="flex text-purple-400 h-full items-center gap-2 justify-end">
          <svg
            width="30"
            height="30"
            viewBox="0 0 30 30"
            className="scale-75"
            fill="none"
            xmlns="http://www.w3.org/2000/svg">
            <path d={ISSUE_STATUS_STEPS[aspectKey]} fill="#673355" />
          </svg>
          {t(`common:${aspectKey}`)}
        </div>
      </div>
      <div className="flex items-center justify-end">
        <ImmuneLink
          href={`${NextJsRoutes.dashboardDevicesIndexPath}?issue=${issue.issueId}`}
          isButton={false}>
          <div className="flex bg-purple-600 hover:bg-purple-300 items-center gap-2 rounded-lg py-0.5 px-3">
            <b>{issue.count}</b>
            <EyeIcon className="h-4" />
          </div>
        </ImmuneLink>
      </div>
      {toggle ? (
        <ChevronDownIcon
          className="h-8 self-center justify-self-end cursor-pointer"
          onClick={toggleDescription}
        />
      ) : (
        <ChevronRightIcon
          className="h-8 self-center justify-self-end cursor-pointer"
          onClick={toggleDescription}
        />
      )}
      {!!toggle && <div className="col-span-4">{description}</div>}
    </div>
  );
};
