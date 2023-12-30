import Headline from 'components/elements/Headlines/Headline';
import IssuesBarGraph from 'components/elements/Issues/IssueBarGraph';
import { IssueTableHeader, IssueTableRow } from 'components/elements/Issues/IssueTable';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLayout from 'components/layouts/dashboard';
import { RISKS_API_URL, useRisks } from 'hooks/issues';
import useTranslation from 'next-translate/useTranslation';
import { useSWRConfig } from 'swr';

const DashboardRisks = (): JSX.Element => {
  const { t } = useTranslation();
  const { data: risksData, isLoading: risksLoading } = useRisks();

  if (risksLoading) return <Spinner />;

  return (
    <SnippetBox role="main">
      <Headline size={5} className="font-semibold">
        {t('dashboard:risks.title')}
      </Headline>
      {risksData?.length ? (
        <>
          <IssuesBarGraph issuesList={risksData} />
          <IssueTableHeader />
          {risksData.map((issue) => (
            <IssueTableRow key={issue.issueId} issue={issue} />
          ))}
        </>
      ) : (
        <p className="text-center text-xl text-gray-400 mt-12 mb-10">
          {t('dashboard:issuesTable.noIssues', {
            issueType: t('dashboard:risks.title').toLowerCase(),
          })}
        </p>
      )}
    </SnippetBox>
  );
};

DashboardRisks.getLayout = function GetLayout(page: React.ReactElement) {
  const { mutate } = useSWRConfig();

  return <DashboardLayout mutateFn={() => mutate(RISKS_API_URL)}>{page}</DashboardLayout>;
};

export default DashboardRisks;
