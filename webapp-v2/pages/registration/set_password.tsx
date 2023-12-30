import Button from 'components/elements/Button/Button';
import Input from 'components/elements/Input/Input';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useSession } from 'hooks/useSession';
import { useChangePasswordUser } from 'hooks/users';
import { useRouter } from 'next/router';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';

export default function SetPassword(): JSX.Element {
  const {
    session: {
      user: { name, email },
    },
  } = useSession();

  const router = useRouter();

  const changePasswordUser = useChangePasswordUser();

  const defaultValues = {
    name,
    email,
    currentPassword: '',
    password: '',
    passwordConfirmation: '',
  };
  const {
    register,
    formState: { errors },
    setError,
    handleSubmit,
  } = useForm({
    defaultValues,
  });

  const changePassword = ({
    password,
    passwordConfirmation,
    currentPassword,
  }: typeof defaultValues): Promise<void> => {
    if (password?.length === 0 || passwordConfirmation?.length === 0) {
      return;
    }
    if (password !== passwordConfirmation) {
      setError('passwordConfirmation', { message: 'New Passwords do not match' });
      return;
    }
    return changePasswordUser.mutate({ password, currentPassword }).then(({ success }) => {
      if (success) {
        router.push(NextJsRoutes.dashboardWelcomePath);
      } else {
        router.push(NextJsRoutes.loginPath);
        toast.error('Something went wrong, setting your password');
      }
    });
  };
  return (
    <DashboardLoggedOut>
      <div className="mt-12 w-[fit-content] self-center flex flex-col space-y-8">
        <span className="text-2xl lg:text-4xl font-bold uppercase">
          You’re successfully registered!
        </span>
        <span className="text-left">
          To continue we need some more information, like a password…
        </span>
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
            />
            <Input
              label="Rerun new Password"
              placeholder="******"
              type="password"
              theme="light"
              {...register(`passwordConfirmation`)}
              errors={[errors?.passwordConfirmation?.message]}
              autoComplete="new-password"
            />
            <Button theme="SUCCESS" type="submit" full={true}>
              Save password
            </Button>
          </div>
        </form>
      </div>
    </DashboardLoggedOut>
  );
}
