import AuthServiceButton from 'components/elements/Button/AuthServiceButton';
import Button from 'components/elements/Button/Button';
import ImmuneLink from 'components/elements/Button/Link';
import Input from 'components/elements/Input/Input';
import Spinner from 'components/elements/Spinner/Spinner';
import ToastContent from 'components/elements/Toast/ToastContent';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useSession } from 'hooks/useSession';
import localforage from 'localforage';
import getConfig from 'next/config';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { useMutation } from 'utils/api';
import { setAuthToken } from 'utils/session';
import { SerializedSession } from 'utils/types';

const { publicRuntimeConfig } = getConfig();

export default function DashboardIndex(): JSX.Element {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const { setSession, logout } = useSession();
  const { isLoading: loginIsLoading, mutate: mutateUseLogin } = useMutation(
    'POST',
    JsRoutesRails.v2_session_path,
  );

  const createSession = (session: SerializedSession): boolean => {
    if (session.memberships?.length > 0) {
      localforage.clear();
      setSession({ ...session, currentMembership: session?.memberships[0] });
      setAuthToken(session?.memberships[0].token);
      return true;
    }
    return false;
  };

  const signInUser = async ({ email = null, password = null, token = null }): Promise<void> => {
    setLoading(true);
    const signInPayload = email && password ? { email, password } : { token };

    try {
      if (loginIsLoading) {
        return;
      }

      const serializedSession = (await mutateUseLogin(signInPayload)) as SerializedSession;

      if (serializedSession?.nextPath) {
        const nextPathLocation = serializedSession.nextPath;

        if (createSession(serializedSession) || nextPathLocation.includes('activate_email?email')) {
          window.location.assign(nextPathLocation);
        } else {
          toast.error('You do not belong to any organisation');
          logout();
        }
      } else {
        setLoading(false);
        toast.error(<ToastContent headline="Login failed" text="Sorry something went wrong" />);
      }
    } catch (error) {
      console.error(error);
      if (error.response && (error.response.status === 401 || error.response.status === 403)) {
        toast.error(
          <ToastContent
            headline="Login failed"
            text={error.response.data.errorMessage || 'Please check your email and password.'}
          />,
        );
      } else {
        toast.error(<ToastContent headline="Login failed" text="Sorry something went wrong" />);
      }
    }
  };

  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm();

  // Login is coming back from oauth
  useEffect(() => {
    if (router.query.token) {
      setLoading(true);
      signInUser({
        token: router.query.token,
      });
    }
  }, [router.query.token]);

  const formHandler = (): void => {
    signInUser({
      email: getValues('email'),
      password: getValues('password'),
    });
  };

  if (loading) return <Spinner />;

  return (
    <DashboardLoggedOut>
      <nav className="absolute top-8 right-8 z-30 flex gap-4">
        {!publicRuntimeConfig.disableRegistration && (
          <ImmuneLink href="/register" theme="WHITE">
            Register
          </ImmuneLink>
        )}
      </nav>
      <h1 className="text-[5rem] text-center mt-12">Sign in</h1>
      <section className="grid items-center gap-y-4 xl:gap-y-0 xl:grid-cols-[1fr,200px,1fr] xl:w-2/3 xl:mt-24 xl:ml-auto xl:mr-auto">
        <form onSubmit={handleSubmit(formHandler)} className="space-y-2">
          <Input
            label="Email Address"
            placeholder="Email Address"
            required={true}
            theme="light"
            type="email"
            errors={[errors.email?.message]}
            {...register('email', {
              required: 'required',
              pattern: {
                value: /\S+@\S+\.\S+/,
                message: 'Entered value does not match email format',
              },
            })}
          />
          <Input
            label="Password"
            placeholder="Password"
            required={true}
            type="password"
            theme="light"
            errors={[errors.password?.message]}
            {...register('password', {
              required: 'required',
            })}
          />
          <Button theme="CTA" full={true} type="submit">
            Sign in with email
          </Button>
          <Link href="/request_new_password" passHref={true}>
            <a className="block text-center">Forgot password? Reset it here...</a>
          </Link>
        </form>
        <div className="flex justify-center h-[110%]">
          <div className="bg-white w-[1px] h-full" />
        </div>
        <div className="space-y-4">
          <AuthServiceButton service="github">Sign in with Github</AuthServiceButton>
          <AuthServiceButton service="google">Sign in with Google</AuthServiceButton>
        </div>
      </section>
    </DashboardLoggedOut>
  );
}
