import { AdjustmentsIcon } from '@heroicons/react/outline';
import { SearchIcon, ChevronLeftIcon } from '@heroicons/react/solid';
import classNames from 'classnames';
import Button from 'components/elements/Button/Button';
import Link from 'components/elements/Button/Link';
import DeviceBox from 'components/elements/DeviceBox/DeviceBox';
import DevicesSidebar from 'components/elements/DeviceBox/DevicesSidebar';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import { getIssueType } from 'components/elements/Issues/Index';
import DeviceTag from 'components/elements/Tag/DeviceTag';
import TagSelect from 'components/elements/TagSelect/TagSelect';
import DashboardLayout from 'components/layouts/dashboard';
import NextJsRoutes from 'generated/NextJsRoutes';
import { getUrlWithFilters, USE_DEVICES_URL } from 'hooks/devices';
import { useInfiniteDeviceLoading } from 'hooks/infinityDevices';
import useIntersectionObserver from 'hooks/intersectionObserver';
import { useSession } from 'hooks/useSession';
import useTranslation from 'next-translate/useTranslation';
import Head from 'next/head';
import { useRouter } from 'next/router';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import { ChangeEvent, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { ApiSrv } from 'types/apiSrv';

const DeviceStates = {
  vulnerable: 'Untrusted',
  trusted: 'Trusted',
  outdated: 'Unresponsive',
  all: 'All',
};

export default function DevicesAll(): JSX.Element {
  const { t } = useTranslation();
  const {
    session: { currentMembership },
  } = useSession();
  const { devices, loadNext, hasNext, isLoading, loadItems } = useInfiniteDeviceLoading();
  const router = useRouter();

  const [stateFilter, setStateFilter] = useState<string>('all');
  const [issueFilter, setIssueFilter] = useState<string>(null);
  const [searchText, setSearchText] = useState('');
  const [filterTags, setFilterTags] = useState<ApiSrv.Tag[]>([]);

  const [filteredDevices, setFilteredDevices] = useState([]);
  const [pushLoad, setPushLoad] = useState(false);

  const ref = useRef<HTMLDivElement | null>(null);
  const currentUrl = useRef(USE_DEVICES_URL);
  const membershipWasSet = useRef(false);
  const entry = useIntersectionObserver(ref, {});

  setTimeout(() => {
    if (entry?.isIntersecting && !pushLoad && devices.length > 0) {
      setPushLoad(true);
    } else if (!entry?.isIntersecting && pushLoad) {
      setPushLoad(false);
    }
  }, 500);

  useEffect(() => {
    if (!devices) return;

    let newFilteredDevices = devices;
    if (searchText) {
      newFilteredDevices = newFilteredDevices.filter(
        (device) =>
          device.name.toLowerCase().includes(searchText.toLocaleLowerCase()) ||
          device.hwid.toLowerCase().includes(searchText) ||
          device.state.toLowerCase().includes(searchText),
      );
    }
    setFilteredDevices(newFilteredDevices);

    setTimeout(() => {
      if (hasNext && entry?.isIntersecting) {
        setPushLoad(false);
      }
    }, 1000);
  }, [devices, searchText]);

  useEffect(() => {
    if (!membershipWasSet.current) return;

    setIssueFilter(null);
    setSearchText('');
    setFilterTags([]);
    setStateFilter('all');
  }, [currentMembership]);

  useEffect(() => {
    if (pushLoad) {
      loadNext();
    }
  }, [pushLoad]);

  useEffect(() => {
    if (!isLoading) {
      setPushLoad(false);
    }
  }, [isLoading]);

  useMemo(() => {
    setStateFilter(router.query.state ? (router.query.state as string) : 'all');
  }, [router.query.state]);

  useMemo(() => {
    if (router.query.issue) setIssueFilter(router.query.issue as string);
  }, [router.query.issue]);

  const issueData = useMemo(() => {
    if (!issueFilter) return {};

    const issueKey = getIssueType(issueFilter);
    const issueKeyTitle = t(`dashboard:${issueKey}.title`);

    return {
      name: t(`${issueKey}:${issueFilter}.slug`, {}, { fallback: `binarly:${issueFilter}.slug` }),
      issueKeyTitle,
      issueType: t(`dashboard:${issueKey}.type`),
      link: NextJsRoutes[`dashboard${issueKeyTitle}Path`],
    };
  }, [issueFilter]);

  useEffect(() => {
    if (stateFilter) {
      const query = stateFilter !== 'all' ? `?state=${stateFilter}` : '';
      router.replace(`${NextJsRoutes.dashboardDevicesIndexPath}${query}`);
    }
  }, [stateFilter]);

  const filterTagsIds = useMemo(() => filterTags.map((tag) => tag.id), [filterTags]);

  useEffect(() => {
    if (currentMembership) membershipWasSet.current = true;

    const newState =
      stateFilter && stateFilter !== 'all' ? (stateFilter as ApiSrv.Device['state']) : null;
    const url = getUrlWithFilters({ stateFilter: newState, filterTags, issueFilter });

    currentUrl.current = url;
    loadItems(currentUrl.current);
  }, [stateFilter, issueFilter, filterTags]);

  const selectedDevices = useContext(SelectedDevicesContext);

  return (
    <>
      <Head>
        <title>immune Guard | Devices</title>
      </Head>
      <div
        className={classNames('grid gap-x-8 transition-all', {
          'grid-cols-[1fr,350px]': selectedDevices.items?.length > 0,
          'grid-cols-[1fr,0] overflow-hidden': selectedDevices.items?.length === 0,
        })}>
        <div>
          {!!issueFilter && (
            <Link href={issueData.link} isButton={false} className="flex -ml-2 mb-4">
              <ChevronLeftIcon className="h-6" />
              <b className="underline">
                {t('dashboard:devices.linkBack', { issueTitle: issueData.issueKeyTitle })}
              </b>
            </Link>
          )}
          <div className="space-y-4">
            <Headline as="h1">{t('dashboard:devices.title')}</Headline>
            <Input
              placeholder={t('dashboard:devices.searchPlaceholder')}
              className="w-3/6"
              icon={<SearchIcon />}
              value={searchText}
              role="search"
              onChange={(e: ChangeEvent<HTMLInputElement>) => setSearchText(e.target.value)}
            />
            {!!issueFilter && (
              <div className="flex justify-between items-center">
                <div className="flex gap-2">
                  <AdjustmentsIcon className="h-6 text-purple-200" />
                  <b>
                    {t('dashboard:devices.filteredByIssue', { issueType: issueData.issueType })}:{' '}
                    {issueData.name}
                  </b>
                </div>
                <Button
                  onClick={() => setIssueFilter(null)}
                  theme="WHITE"
                  className="border-purple-300 w-28	self-right py-1 px-1 my-0">
                  {t('dashboard:devices.resetFilter')}
                </Button>
              </div>
            )}
            <div className="mt-16 space-y-4">
              <TagSelect
                selectedKey={stateFilter}
                options={DeviceStates}
                onSelect={(key) => setStateFilter(key)}
              />
            </div>
            {filterTags.length > 0 && (
              <div className="flex space-x-2 items-center">
                <b>{t('dashboard:devices.filterByTags')}:</b>
                {filterTags.map((filterTag) => (
                  <DeviceTag
                    key={filterTag.id}
                    tag={filterTag}
                    remove={true}
                    onClick={() => {
                      const newTags = filterTags.filter((t) => t.id !== filterTag.id);
                      setFilterTags(newTags);
                    }}
                  />
                ))}
              </div>
            )}
          </div>
          <div className="mt-14 space-y-4" role="table">
            <div className="flex justify-between items-end">
              <b className="text-purple-500">{`${filteredDevices?.length} Results`}</b>
              {filteredDevices?.length ? (
                <Button
                  onClick={() =>
                    selectedDevices.dispatch({
                      type: DeviceActionType.SELECT_ALL,
                      devices: filteredDevices,
                    })
                  }
                  className="w-28 self-right py-1 px-1 my-0">
                  {t('dashboard:devices.selectAll')}
                </Button>
              ) : null}
            </div>
            {filteredDevices?.map((device) => (
              <DeviceBox
                key={device.id}
                device={device}
                selected={!!selectedDevices.items?.find((item) => item.id === device.id)}
                onTagClick={(tag) => {
                  if (!filterTagsIds.includes(tag.id)) {
                    setFilterTags([...filterTags, tag]);
                  }
                }}
                onCheckboxClick={(device) =>
                  selectedDevices.dispatch({ type: DeviceActionType.TOGGLE, device })
                }
              />
            ))}
          </div>
          <div className="flex mt-8 justify-center">
            {hasNext ? (
              <div ref={ref}>
                <Button onClick={() => loadNext()} theme="CTA">
                  {isLoading
                    ? t('dashboard:devices.loading')
                    : t('dashboard:devices.loadNextDevices')}
                </Button>
              </div>
            ) : (
              <div>{t('dashboard:devices.allLoaded')}</div>
            )}
          </div>
        </div>
        <DevicesSidebar />
      </div>
    </>
  );
}

DevicesAll.getLayout = function getLayout(page: React.ReactElement) {
  return <DashboardLayout className="relative">{page}</DashboardLayout>;
};
