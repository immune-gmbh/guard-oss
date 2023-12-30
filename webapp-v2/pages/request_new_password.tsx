import Button from 'components/elements/Button/Button';
import Input from 'components/elements/Input/Input';
import Spinner from 'components/elements/Spinner/Spinner';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import getConfig from 'next/config';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { authenticatedJsonapiRequest } from 'utils/fetcher';

export default function DashboardIndex(): JSX.Element {
  const [loading, setLoading] = useState(false);
  const [send, setSend] = useState(false);
  const { publicRuntimeConfig } = getConfig();

  const {
    register,
    handleSubmit,
    getValues,
    formState: { errors },
  } = useForm();

  const formHandler = async (): Promise<void> => {
    setLoading(true);
    setSend(true);

    const data = (await authenticatedJsonapiRequest(
      `${publicRuntimeConfig.hosts.authSrv}${JsRoutesRails.v2_password_reset_index_path()}`,
      {
        method: 'POST',
        body: JSON.stringify({ email: getValues('email') }),
      },
    )) as any;

    setLoading(false);
    if (data.ok) {
      toast.success('You will receive an email with information how to reset your password');
    } else {
      toast.error('Something went wrong, please try again (later).');
    }
  };

  if (loading) return <Spinner />;

  return (
    <DashboardLoggedOut>
      <h1 className="text-[5rem] text-center mt-12">Reset-Password</h1>
      <section className="grid items-center xl:w-2/3 xl:mt-24 xl:ml-auto xl:mr-auto">
        {!send ? (
          <form onSubmit={handleSubmit(formHandler)} className="space-y-2" data-testid="form">
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
            <Button theme="CTA" full={true} type="submit">
              Reset password
            </Button>
          </form>
        ) : (
          <span className="text-center">
            You will receive an email with information how to reset your password
          </span>
        )}
      </section>
    </DashboardLoggedOut>
  );
}
