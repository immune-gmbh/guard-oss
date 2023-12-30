import { ChevronDownIcon } from '@heroicons/react/outline';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import ReactECharts from 'echarts-for-react';
import NextJsRoutes from 'generated/NextJsRoutes';
import { IRisksRanking } from 'hooks/dashboard';
import useTranslation from 'next-translate/useTranslation';
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { Aspect, aspectByIssueId } from 'utils/issues';

const defaultGraphOptions = {
  grid: { left: 20, top: 30, right: 0, bottom: 35 },
  xAxis: {
    type: 'category',
    data: [],
    axisLabel: {
      interval: 0,
      overflow: 'break',
      fontSize: 10,
      width: 56,
    },
    axisTick: { show: false },
    axisLine: { show: false },
  },
  yAxis: {
    type: 'value',
    splitNumber: 3,
    axisLabel: { showMinLabel: false },
    offset: -2,
    axisLine: { onZero: false },
  },
  textStyle: {
    fontFamily: 'Eurostile',
  },
  series: [
    {
      type: 'bar',
      data: [],
      barWidth: 20,
      itemStyle: {
        borderRadius: 5,
        color: '#c6b3bf',
      },
    },
  ],
};

const RisksRanking = ({ risks }: { risks: IRisksRanking[] }): JSX.Element => {
  const { t } = useTranslation();
  const [option, setOption] = useState({});

  const topFiveList = risks
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
    .map((risk) => {
      return (
        <li key={risk.issueId} className="odd:bg-gray-100 py-2	px-3 font-semibold">
          {t(`risks:${risk.issueId}.slug`, {}, { fallback: `incidents:${risk.issueId}.slug` })}
        </li>
      );
    });

  const risksByAspect = risks.reduce((aspectCounter, risk) => {
    const aspect = aspectByIssueId(risk.issueId);
    aspectCounter[aspect] = (aspectCounter[aspect] || 0) + risk.count;
    return aspectCounter;
  }, {});

  const aspectData = Object.values(Aspect).reduce((aspectCounter, aspect) => {
    aspectCounter[aspect] = aspectCounter[aspect] || 0;
    aspectCounter[aspect] += risksByAspect[aspect] || 0;

    return aspectCounter;
  }, {});

  const graphOption = {
    ...defaultGraphOptions,
    xAxis: {
      ...defaultGraphOptions.xAxis,
      data: Object.keys(aspectData).map((key) => t(`common:${key}`)),
    },
    series: [
      {
        ...defaultGraphOptions.series[0],
        data: Object.values(aspectData),
      },
    ],
  };

  const noRisks = risks.length == 0;

  useEffect(() => {
    setOption(graphOption);
  }, []);

  return (
    <SnippetBox className="min-h-[10rem] p-6 min-w-[58%] flex-1 bg-white justify-between">
      <div>
        <div className="flex justify-between items-center mb-4 gap-2">
          <Headline size={5} className="font-semibold">
            {t('dashboard:risksRanking.title')}
          </Headline>
          <Button
            theme="WHITE"
            disabled={true}
            className="py-0 pr-1.5 pl-2.5 h-6 rounded-lg text-slate-900 border-slate-400">
            <span className="text-xs whitespace-pre">{t('dashboard:timestamp')}</span>
            <ChevronDownIcon className="h-3" />
          </Button>
        </div>
        <ul role="list">{topFiveList}</ul>
        {!noRisks && (
          <Link href={NextJsRoutes.dashboardRisksPath} passHref>
            <a>
              <Button className="w-28	self-center py-1 px-1 my-4 mx-auto">
                {t('dashboard:risksRanking.showAll')}
              </Button>
            </a>
          </Link>
        )}
      </div>
      {noRisks && (
        <p className="text-center text-xl text-gray-400 whitespace-pre">
          {t('dashboard:risksRanking.noRisks')}
        </p>
      )}
      <hr />
      {!noRisks && (
        <div className="w-full min-h-[12rem]">
          <p className="font-semibold text-purple-500 mt-4">
            {t('dashboard:risksRanking.byTrustchain')}
          </p>
          <ReactECharts option={option} style={{ height: '160px' }} />
        </div>
      )}
    </SnippetBox>
  );
};

export default RisksRanking;
