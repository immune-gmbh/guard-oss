import ReactECharts from 'echarts-for-react';
import { IIssueRanking } from 'hooks/dashboard';
import useTranslation from 'next-translate/useTranslation';
import { useEffect, useState } from 'react';
import { aspectByIssueId } from 'utils/issues';

const COLOR = {
  'purple-5': '#ECE6EA',
  'purple-4': '#C6B3BF',
  'purple-3': '#8D667F',
  'purple-2': '#673355',
  'purple-1': '#54193F',
  'purple-0': '#410028',
};

const colorPriority = [1, 2, 4, 3, 5, 0];

const defaultGraphOptions = {
  grid: { left: 0, top: 0, right: 0, bottom: 10 },
  legend: {
    type: 'scroll',
    left: 0,
    bottom: 0,
    icon: 'circle',
    itemWidth: 10,
    selectedMode: false,
  },
  textStyle: {
    fontFamily: 'Eurostile',
  },
  xAxis: {
    type: 'value',
    show: false,
  },
  yAxis: {
    type: 'category',
    data: ['Types'],
    show: false,
  },
  series: [],
};

const defaultSeriesOptions = {
  type: 'bar',
  stack: 'total',
  barWidth: 30,
  itemStyle: {
    borderWidth: 2,
    borderColor: '#fff',
  },
  barMinHeight: 20,
};

const IssuesBarGraph = ({ issuesList }: { issuesList: IIssueRanking[] }): JSX.Element => {
  const { t } = useTranslation();
  const [option, setOption] = useState({});

  const graphData = issuesList.reduce(
    (aspectObj: { [key: string]: number }, issue: IIssueRanking) => {
      const aspectKey = aspectByIssueId(issue.issueId);
      aspectObj[aspectKey] = (aspectObj[aspectKey] || 0) + issue.count;
      return aspectObj;
    },
    {},
  );

  const graphDataWithCount = Object.entries(graphData).filter(
    (aspect: [string, number]) => aspect[1] > 0,
  );

  const seriesOptions = graphDataWithCount.map(([key, count]: [string, number], index, entries) => {
    const seriesCount = entries.length;
    const borderRadius = [
      index == 0 ? 30 : 0,
      index == seriesCount - 1 ? 30 : 0,
      index == seriesCount - 1 ? 30 : 0,
      index == 0 ? 30 : 0,
    ];
    const activeColors = colorPriority.slice(-seriesCount).sort();

    return {
      ...defaultSeriesOptions,
      name: `${t(`common:${key}`)} [${count}]`,
      data: [count],
      itemStyle: {
        ...defaultSeriesOptions.itemStyle,
        color: COLOR[`purple-${activeColors[index]}`],
        borderRadius,
      },
    };
  });

  const xAxisMax = Object.values(graphData).reduce(
    (total: number, curr: number) => total + curr,
    0,
  );

  const edgeAdjustment = () => {
    const lastEntry = graphDataWithCount[graphDataWithCount.length - 1];
    return (lastEntry?.[1] || 0) < xAxisMax * 0.01 ? xAxisMax * 1.02 : xAxisMax;
  };

  const graphOption = {
    ...defaultGraphOptions,
    xAxis: {
      ...defaultGraphOptions.xAxis,
      max: edgeAdjustment(),
    },
    series: seriesOptions,
  };

  useEffect(() => {
    setOption(graphOption);
  }, [issuesList]);

  return <ReactECharts option={option} style={{ height: '80px' }} />;
};

export default IssuesBarGraph;
