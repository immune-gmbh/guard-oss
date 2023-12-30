import Button from 'components/elements/Button/Button';
import Link from 'components/elements/Button/Link';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import ToastContent from 'components/elements/Toast/ToastContent';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useResendActivationEmailUser } from 'hooks/users';
import router, { useRouter } from 'next/dist/client/router';
import Image from 'next/image';
import ProtectedServerSvg from 'public/img/protected_server.svg';
import React from 'react';
import { toast } from 'react-toastify';

export default function RegistrationSend(): JSX.Element {
  const { query } = useRouter();

  const { mutate, isLoading } = useResendActivationEmailUser();

  if (!query.email) {
    router.push(NextJsRoutes.loginPath);
    toast.error('No Email entered.');
    return null;
  }

  return (
    <DashboardLoggedOut>
      <nav className="absolute top-8 right-8 z-30 flex gap-4">
        <Link href={NextJsRoutes.loginPath} theme="WHITE">
          Login
        </Link>
      </nav>
      <h1 className="text-[3rem] md:text-[5rem] text-center mt-12">Confirm your email address</h1>
      <div className="text-black flex flex-col items-center space-y-4 p-8 lg:max-w-[50%] md:p-[20%] lg:p-[10%] mx-auto text-center bg-white md:bg-transparent md:bg-hexagon-bg bg-contain bg-center bg-no-repeat ">
        <div className="font-bold text-xl text-white">We send an email to {query.email}</div>
        <div>
          <Image src={ProtectedServerSvg} />
        </div>
        <span className="text-white">
          Please confirm your email address by clicking the link we just sent to your inbox
        </span>
        {!isLoading && (
          <Button
            theme="CTA"
            onClick={() =>
              mutate({ email: query.email }).then(() =>
                toast.success(<ToastContent headline="Email resend" text="Check your inbox." />),
              )
            }>
            Resend verification email
          </Button>
        )}
        {isLoading && <LoadingSpinner />}
      </div>
    </DashboardLoggedOut>
  );
}
