import { ChevronDownIcon } from '@heroicons/react/outline';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import ReactECharts from 'echarts-for-react';
import { IDeviceStats } from 'hooks/dashboard';
import { useOrganisation } from 'hooks/organisations';
import { useSubscription } from 'hooks/subscriptions';
import _ from 'lodash';
import useTranslation from 'next-translate/useTranslation';
import { useEffect, useRef, useState } from 'react';

const defaultGraphOptions = {
  grid: { left: 0, top: 0, right: 0, bottom: 30 },
  legend: [
    {
      data: [],
      top: 'bottom',
      icon: 'circle',
      itemWidth: 8,
    },
  ],
  textStyle: {
    fontFamily: 'Eurostile',
    fontSize: 16,
  },
  series: [
    {
      type: 'graph',
      layout: 'force',
      force: {
        repulsion: 120,
        layoutAnimation: false,
        edgeLength: 5,
      },
      label: { show: true },
      categories: [],
      zoom: 1.5,
      top: 'middle',
      left: 'center',
      data: [],
      links: [],
      emphasis: {
        focus: 'self',
        label: {
          position: 'top',
          show: true,
          color: '#000',
        },
      },
    },
  ],
};

const DeviceStatus = ({
  organisationId,
  deviceStats,
}: {
  organisationId: string;
  deviceStats: IDeviceStats;
}): JSX.Element => {
  const { t } = useTranslation();
  const [option, setOption] = useState({});
  const widthRef = useRef(null);
  const statsRef = useRef(null);
  const [isLoading, setIsLoading] = useState(true);

  const { data: organisation } = useOrganisation({ id: organisationId });
  const { data: subscription } = useSubscription(organisation?.subscription?.id);

  const data = parseBubbleValues(deviceStats);
  const links = data.map((node) => node.link);
  const edgeLength = data.map((d) => d.link.value);
  const calcRepulsion = (): number => {
    const graphWidth = widthRef.current?.offsetWidth || 0;
    if (window.innerWidth <= 1112) return 0;

    if (graphWidth > 550) return -100;
    if (graphWidth > 400) return -50;
    return 0;
  };

  const repulsion = calcRepulsion();

  const graphOption = {
    ...defaultGraphOptions,
    legend: [
      {
        ...defaultGraphOptions.legend[0],
        data: data.map((node) => ({ ...node, itemStyle: { color: node.color } })),
      },
    ],
    series: [
      {
        ...defaultGraphOptions.series[0],
        categories: data,
        force: {
          ...defaultGraphOptions.series[0].force,
          edgeLength,
          repulsion,
        },
        data: data.map((node) => ({
          ...node,
          symbolSize: Math.max(node.size, 10),
          itemStyle: {
            color: node.color,
          },
          label: {
            fontWeight: 'bold',
            show: true,
            fontSize: 18,
            formatter(d: { value: number }) {
              return d.value;
            },
            ...(node.size < 20 && { color: '#00000000' }),
          },
        })),
        links: links.length > 1 ? links : [],
      },
    ],
  };

  useEffect(() => {
    if (_.isEqual(statsRef.current, deviceStats)) return;

    setIsLoading(true);
    setTimeout(() => {
      setOption({ ...graphOption });
      setIsLoading(false);
    }, 600);

    statsRef.current = deviceStats;
  }, [deviceStats]);

  return (
    <SnippetBox className="p-6 min-w-[40%] flex-1 bg-white h-[30rem] 3xl:min-h-[40rem]">
      <div className="flex justify-between items-center gap-4">
        <Headline size={5} className="font-semibold">
          {t('dashboard:deviceStatus.title')}
        </Headline>
        <Button
          theme="WHITE"
          disabled={true}
          className="py-0 pr-1.5 pl-2.5 h-6 rounded-lg text-slate-900 border-slate-400">
          <span className="text-xs whitespace-pre">{t('dashboard:timestamp')}</span>
          <ChevronDownIcon className="h-3" />
        </Button>
      </div>
      {subscription && (
        <div role="status">
          <span className="font-semibold">{`${subscription.currentDevicesAmount} / ${subscription.maxDevicesAmount}`}</span>{' '}
          {t('dashboard:deviceStatus.devices')}
        </div>
      )}
      <div
        className="w-full h-[calc(100%_-_6.5rem)] flex grow items-center justify-center relative"
        ref={widthRef}>
        {isLoading && <LoadingSpinner />}
        {!isLoading && (
          <ReactECharts
            style={{
              height: 'min(32rem, 100%)',
              width: '100%',
            }}
            option={option}
            notMerge={true}
          />
        )}
      </div>
    </SnippetBox>
  );
};

export default DeviceStatus;

type BubbleOptions = {
  color: string;
  name: string;
  id: string;
  x: number;
  y: number;
  value: number;
  size: number;
  link: {
    source: string;
    target: string;
    value: number;
  };
};

export const parseBubbleValues = (stats: IDeviceStats): Array<BubbleOptions> => {
  const maxValue = Math.max(...(Object.values(stats || {}) as number[]));
  const setSize = (count: number): number => MAX_BUBBLE_SIZE * Math.sqrt(count / (maxValue || 1));

  return Object.entries(stats || {})
    .filter((stat: [string, number]) => stat[1] > 0)
    .sort((a: [string, number], b: [string, number]) => b[1] - a[1])
    .map(([key, count]: [string, number], index, entries) => {
      return {
        ...bubbleNode[key],
        ...bubbleCoordinates[index],
        value: count,
        size: setSize(count),
        link: {
          source: bubbleNode[entries[0][0]].id,
          target: key,
          value: Math.max(setSize(count) / 4, 32),
        },
      };
    });
};

const MAX_BUBBLE_SIZE = 120;

const bubbleNode = {
  numTrusted: {
    color: '#e0e9c7',
    name: 'Secure',
    id: 'numTrusted',
  },
  numAtRisk: {
    color: '#8d667f',
    name: 'At Risk',
    id: 'numAtRisk',
  },
  numUnresponsive: {
    color: '#c6b3bf',
    name: 'Unresponsive',
    id: 'numUnresponsive',
  },
  numWithIncident: {
    color: '#ff193c',
    name: 'Incident',
    id: 'numWithIncident',
  },
};

const bubbleCoordinates = [
  { x: 0, y: 0 },
  { x: 90, y: -45 },
  { x: -60, y: 60 },
  { x: 60, y: 45 },
];
