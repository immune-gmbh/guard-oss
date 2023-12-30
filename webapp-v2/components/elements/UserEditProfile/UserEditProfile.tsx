import Actionbar from 'components/elements/Actionbar/Actionbar';
import { SERVICE_ICONS } from 'components/elements/Button/AuthServiceButton';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import ConfirmModal from 'components/elements/Modal/ConfirmModal';
import { useSession } from 'hooks/useSession';
import { useChangePasswordUser, useDeleteUser, useUpdateUser } from 'hooks/users';
import { useForm, useFormState } from 'react-hook-form';
import { toast } from 'react-toastify';

const UserEditProfile: React.FC = () => {
  const {
    session: {
      id,
      user: { name, email, ...userRest },
      ...sessionRest
    },
    setSession,
    logout,
  } = useSession();

  const updateUser = useUpdateUser();
  const changePasswordUser = useChangePasswordUser();
  const deleteUser = useDeleteUser();

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
    control,
    reset,
    setValue,
  } = useForm({
    defaultValues,
  });
  const { dirtyFields } = useFormState({
    control,
  });

  const changePassword = ({
    password,
    passwordConfirmation,
    currentPassword,
  }: typeof defaultValues): Promise<void> => {
    if (
      password?.length === 0 ||
      passwordConfirmation?.length === 0 ||
      currentPassword?.length === 0
    ) {
      return;
    }
    if (password !== passwordConfirmation) {
      setError('passwordConfirmation', { message: 'New Passwords do not match' });
      return;
    }
    return changePasswordUser.mutate({ password, currentPassword }).then(({ errors }) => {
      if (errors && errors.length > 0) {
        toast.error(errors[0]?.title || 'Failed to change password');
      } else {
        toast.success('Sucessfully changed passwords.');
        setValue('password', '');
        setValue('currentPassword', '');
        setValue('passwordConfirmation', '');
        reset({}, { keepValues: true });
      }
    });
  };

  const changeProfile = async (data: typeof defaultValues) => {
    if (!dirtyFields.email && !dirtyFields.name) {
      return;
    }

    const updateResponse = (await updateUser.mutate({ ...data, id: userRest.id })) as any;

    let hasErrors = false;
    Array.from(updateResponse.errors || [])
      .filter(({ id: fieldName }) => ['name', 'email'].includes(fieldName))
      .map(({ id: fieldName, title: error }: { id: string; title: string }) => {
        hasErrors = true;
        setError(fieldName as 'name' | 'email', { message: error });
      });
    if (!hasErrors) {
      toast.success('Sucessfully changed Profile.');
      setSession({ id, ...sessionRest, user: { email, name, ...userRest } });
      reset({ email: email as string, name: name as string });
    }
  };

  const onUpdateUser = (data: typeof defaultValues): void => {
    changePassword(data);
    changeProfile(data);
  };

  const onDeleteUser = (): void => {
    deleteUser.mutate({ id: userRest.id }).then(() => {
      logout();
    });
  };

  return (
    <form role="form" onSubmit={handleSubmit(onUpdateUser)}>
      <div className="space-y-6 max-w-[50ch]">
        <Headline size={4}>Profile information</Headline>
        <div className="">
          <Input
            label="Name"
            placeholder="Name"
            {...register(`name`, {
              required: { value: true, message: 'Field is required' },
            })}
            errors={[errors?.name?.message]}
            autoComplete="username"
          />
          <Input
            label="Email"
            placeholder="Email"
            {...register(`email`, {
              required: { value: true, message: 'Field is required' },
            })}
            errors={[errors?.email?.message]}
            autoComplete="email"
          />
        </div>
        <hr />
        {[
          { service: 'github', value: userRest.authenticatedGithub },
          { service: 'google', value: userRest.authenticatedGoogle },
        ].map((service) => {
          if (service.value) {
            return (
              <div className="flex items-center">
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 18 20"
                  fill="currentColor"
                  className="mr-3 text-opacity-50 transform">
                  <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d={SERVICE_ICONS[service.service]}></path>
                </svg>
                <span>Authenticated with {service.service}</span>
              </div>
            );
          }
          return null;
        })}
        {userRest.hasPassword && (
          <div>
            <Input
              label="Current Password"
              placeholder="******"
              type="password"
              {...register(`currentPassword`)}
              errors={[errors?.currentPassword?.message]}
              autoComplete="current-password"
            />
            <Input
              label="New Password"
              placeholder="******"
              type="password"
              {...register(`password`)}
              errors={[errors?.password?.message]}
              autoComplete="new-password"
            />
            <Input
              label="Rerun new Password"
              placeholder="******"
              type="password"
              {...register(`passwordConfirmation`)}
              errors={[errors?.passwordConfirmation?.message]}
              autoComplete="new-password"
            />
          </div>
        )}
        <ConfirmModal
          onConfirm={onDeleteUser}
          text="Do you really want to delete your account?"
          TriggerComponent={(props) => (
            <Button theme="SECONDARY" {...props}>
              Delete Account
            </Button>
          )}
        />
        <Actionbar>
          <div></div>
          <Button
            data-qa="submit-changes"
            theme="SUCCESS"
            onClick={handleSubmit(onUpdateUser)}
            disabled={!Object.values(dirtyFields).some((f) => f)}>
            Save changes
          </Button>
        </Actionbar>
      </div>
    </form>
  );
};
export default UserEditProfile;
