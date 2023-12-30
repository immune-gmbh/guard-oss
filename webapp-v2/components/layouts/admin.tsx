import classNames from 'classnames';
import AdminSidebarNavigation from 'components/elements/Navigations/Admin/AdminSidebarNavigation';
import Head from 'next/head';
import React from 'react';

type IAdminLayoutProps = React.HTMLProps<HTMLDivElement>;

export default function AdminLayout({
  className,
  children,
  ...rest
}: IAdminLayoutProps): JSX.Element {
  return (
    <>
      <Head>
        <title>immune Guard | Admin</title>
      </Head>
      <div className="w-full flex items-stretch bg-cell-corner bg-no-repeat bg-right-bottom">
        <AdminSidebarNavigation />
        <div className="w-content min-h-screen flex flex-col flex-1">
          <div className={classNames('p-8 flex-1', className)} {...rest}>
            {children}
          </div>
          <div id="actionbar-slot" className="sticky left-0 bottom-0 bg-purple-500"></div>
        </div>
      </div>
    </>
  );
}
