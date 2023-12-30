import Headline from 'components/elements/Headlines/Headline';
import IssuesBarGraph from 'components/elements/Issues/IssueBarGraph';
import { IssueTableHeader, IssueTableRow } from 'components/elements/Issues/IssueTable';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import { INCIDENTS_API_URL, useIncidents } from 'hooks/issues';
import useTranslation from 'next-translate/useTranslation';
import { useSWRConfig } from 'swr';

const DashboardIncidents = (): JSX.Element => {
  const { t } = useTranslation();
  const { data: incidentsData, isLoading: incidentsLoading } = useIncidents();

  if (incidentsLoading) return <Spinner />;

  return (
    <SnippetBox role="main">
      <Headline size={5} className="font-semibold">
        {t('dashboard:incidents.title')}
      </Headline>
      {incidentsData?.length ? (
        <>
          <IssuesBarGraph issuesList={incidentsData} />
          <IssueTableHeader />
          {incidentsData.map((issue) => (
            <IssueTableRow key={issue.issueId} issue={issue} />
          ))}
        </>
      ) : (
        <p className="text-center text-xl text-gray-400 mt-12 mb-10">
          {t('dashboard:issuesTable.noIssues', {
            issueType: t('dashboard:incidents.title').toLowerCase(),
          })}
        </p>
      )}
    </SnippetBox>
  );
};

DashboardIncidents.getLayout = function GetLayout(page: React.ReactElement) {
  const { mutate } = useSWRConfig();

  return <DashboardLayout mutateFn={() => mutate(INCIDENTS_API_URL)}>{page}</DashboardLayout>;
};

export default DashboardIncidents;
