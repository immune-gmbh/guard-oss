import Headline from 'components/elements/Headlines/Headline';
import NextJsRoutes from 'generated/NextJsRoutes';
import Head from 'next/head';
import Link from 'next/link';
import { Fragment } from 'react';
import { UrlObject } from 'url';

interface IBreadcrumbNavigation {
  crumbs: {
    title: string;
    path?: string | UrlObject;
    component?: JSX.Element;
  }[];
}

export default function BreadcrumbNavigation({ crumbs }: IBreadcrumbNavigation): JSX.Element {
  const title = crumbs.map((crumb) => crumb.title).join(' | ');
  const lastCrumb = crumbs[crumbs.length - 1];

  return (
    <>
      <Head>
        <title>immune Guard | {title}</title>
      </Head>
      <nav>
        <span className="uppercase text-purple-500 font-bold text-sm">
          <Link href={NextJsRoutes.dashboardDevicesIndexPath}>Devices</Link>
        </span>
        <ul className="flex flex-row items-center gap-4">
          {crumbs.slice(0, crumbs.length - 1).map((crumb) => {
            return (
              <Fragment key={crumb.title}>
                <li className="text-3xl font-normal text-gray-500">
                  {crumb.component ? crumb.component : <Link href={crumb.path}>{crumb.title}</Link>}
                </li>
                <li>&mdash;</li>
              </Fragment>
            );
          })}
          <li>
            {lastCrumb.component ? (
              lastCrumb.component
            ) : (
              <Headline as="h1" size={4} bold={true} className="text-purple-500">
                <Link href={lastCrumb.path}>{lastCrumb.title}</Link>
              </Headline>
            )}
          </li>
        </ul>
      </nav>
    </>
  );
}
