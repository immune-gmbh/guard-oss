import Button from 'components/elements/Button/Button';
import Input from 'components/elements/Input/Input';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import NextJsRoutes from 'generated/NextJsRoutes';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import getConfig from 'next/config';
import { useRouter } from 'next/router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { authenticatedJsonapiRequest } from 'utils/fetcher';

export default function SetPassword(): JSX.Element {
  const router = useRouter();
  const [updated, setUpdated] = useState(false);

  const defaultValues = {
    currentPassword: '',
    password: '',
    passwordConfirmation: '',
  };
  const {
    register,
    formState: { errors },
    setError,
    handleSubmit,
    getValues,
  } = useForm({
    defaultValues,
  });

  const changePassword = async ({
    password,
    passwordConfirmation,
  }: typeof defaultValues): Promise<void> => {
    const { publicRuntimeConfig } = getConfig();
    setUpdated(true);

    if (password?.length === 0 || passwordConfirmation?.length === 0) {
      return;
    }
    if (password !== passwordConfirmation) {
      setError('passwordConfirmation', { message: 'New Passwords do not match' });
      return;
    }

    try {
      await authenticatedJsonapiRequest(
        `${publicRuntimeConfig.hosts.authSrv}${JsRoutesRails.v2_password_reset_path({
          id: router.query.resetToken,
        })}`,
        {
          method: 'PATCH',
          body: JSON.stringify({ password: getValues('password') }),
        },
      );
      toast.success('Your password was successfully updated', {
        onClose: () => router.push(NextJsRoutes.loginPath),
      });
    } catch {
      toast.success('Something went wrong, when we tried to update your password', {
        onClose: () => router.push(NextJsRoutes.loginPath),
      });
    }
  };

  return (
    <DashboardLoggedOut>
      <div className="mt-12 self-center flex flex-col space-y-8 w-full items-center">
        <span className="text-2xl lg:text-4xl font-bold uppercase">Reset password</span>
        <span className="text-left">You can now set a new password.</span>
        <form onSubmit={handleSubmit(changePassword)} className="self-center w-1/2">
          <div className="space-y-6">
            <Input
              label="New Password"
              placeholder="******"
              type="password"
              theme="light"
              {...register(`password`)}
              errors={[errors?.password?.message]}
              autoComplete="new-password"
              readOnly={updated}
            />
            <Input
              label="Rerun new Password"
              placeholder="******"
              type="password"
              theme="light"
              {...register(`passwordConfirmation`)}
              errors={[errors?.passwordConfirmation?.message]}
              autoComplete="new-password"
              readOnly={updated}
            />
            <Button theme="SUCCESS" type="submit" full={true} disabled={updated}>
              Set password
            </Button>
          </div>
        </form>
      </div>
    </DashboardLoggedOut>
  );
}
