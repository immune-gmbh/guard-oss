import AuthServiceButton from 'components/elements/Button/AuthServiceButton';
import Button from 'components/elements/Button/Button';
import Link from 'components/elements/Button/Link';
import Input from 'components/elements/Input/Input';
import ToastContent from 'components/elements/Toast/ToastContent';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import NextJsRoutes from 'generated/NextJsRoutes';
import { useCreateUser } from 'hooks/users';
import { useRouter } from 'next/router';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { SerializedUser, Response } from 'utils/types';

export default function SignIn(): JSX.Element {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm();
  const createUser = useCreateUser();
  const formHandler = async (): Promise<void> => {
    try {
      const user = (await createUser.mutate({
        name: getValues('name'),
        email: getValues('email'),
        password: getValues('password'),
      })) as Response<SerializedUser>;

      if (!user || user?.errors?.length) {
        toast.error(
          <ToastContent
            headline="Registration failed"
            text="Please check your email and password."
          />,
        );
      } else {
        router.push(`${NextJsRoutes.registrationActivateEmailPath}?email=${getValues('email')}`);
      }
    } catch (e) {
      toast.error(
        <ToastContent
          headline="Registration failed"
          text="Please check your email and password."
        />,
      );
    }
  };

  return (
    <DashboardLoggedOut>
      <nav className="absolute top-8 right-8 z-30 flex gap-4">
        <Link href="/login" theme="GHOST-WHITE">
          Sign in
        </Link>
      </nav>
      <h1 className="text-[5rem] text-center mt-12">Register</h1>
      <section className="grid items-center gap-y-4 xl:gap-y-0 xl:grid-cols-[1fr,200px,1fr] xl:w-2/3 xl:mt-24 xl:ml-auto xl:mr-auto">
        <form onSubmit={handleSubmit(formHandler)} className="space-y-2">
          <Input
            label="Name"
            placeholder="Name"
            required={true}
            theme="light"
            errors={[errors.email?.message]}
            {...register('name', {
              required: 'required',
            })}
          />
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
            Sign up with email
          </Button>
        </form>
        <div className="flex justify-center h-[110%]">
          <div className="bg-white w-[1px] h-full" />
        </div>
        <div className="space-y-4">
          <AuthServiceButton service="github">Sign up with Github</AuthServiceButton>
          <AuthServiceButton service="google">Sign up with Google</AuthServiceButton>
        </div>
      </section>
    </DashboardLoggedOut>
  );
}
