import classNames from 'classnames';
import LiteToWhiteAnchor from 'components/elements/Anchor/LiteToWhiteAnchor';
import MenuListIcon from 'components/elements/Icons/MenuListIcon';
import Link from 'next/link';
import { useRouter } from 'next/router';
import React from 'react';
import { UrlObject } from 'url';

export interface INavigationLinkProps {
  label: string;
  href: string | UrlObject;
  Icon?: (props: React.ComponentProps<'svg'>) => JSX.Element;
  hide?: boolean;
}

const NavigationLink: React.FC<INavigationLinkProps> = ({ label, href, Icon = MenuListIcon }) => {
  const router = useRouter();

  let link = href;
  let active = false;
  let queryActive = false;
  let hasQuery = false;
  if (typeof href === 'object') {
    link = href.pathname;
    if (href.query) {
      hasQuery = true;
      queryActive = JSON.stringify(href.query) === JSON.stringify(router.query);
    }
  }

  if (queryActive) {
    if (link === router.pathname) {
      active = true;
    }
  } else {
    active = link === router.pathname && !hasQuery && Object.keys(router.query).length === 0;
  }

  return (
    <li
      className={classNames('flex items-center text-white text-opacity-60', {
        'text-opacity-100': active,
      })}>
      <Icon height="20" className="pr-4 ml-[0.125em]" />
      <Link href={href} passHref>
        <LiteToWhiteAnchor>{label}</LiteToWhiteAnchor>
      </Link>
    </li>
  );
};
export default NavigationLink;
