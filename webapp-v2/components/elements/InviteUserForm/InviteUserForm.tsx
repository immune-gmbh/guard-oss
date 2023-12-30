import Button from 'components/elements/Button/Button';
import Input from 'components/elements/Input/Input';
import Select from 'components/elements/Select/Select';
import ToastContent from 'components/elements/Toast/ToastContent';
import { useSession } from 'hooks/useSession';
import { useInviteUser } from 'hooks/users';
import React, { useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'react-toastify';

interface IInviteUserFormProps {
  showNameInput?: boolean;
}

const InviteUserForm: React.FC<IInviteUserFormProps> = ({ showNameInput = false }) => {
  const inviteUser = useInviteUser();
  const {
    session: { currentMembership },
  } = useSession();
  const defaultNameValue = showNameInput ? '' : 'Anonymous';

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm({
    defaultValues: {
      users: [{ name: defaultNameValue, email: '', role: 'user' }],
    },
  });
  const { fields, append } = useFieldArray({
    control,
    name: 'users',
  });
  const [completedRows, setCompletedRows] = useState([]);
  const onSubmit = (data): void => {
    data.users.map((user, index) => {
      if (completedRows.includes(index)) {
        return;
      }
      return (
        inviteUser
          .mutate({ membership: { ...user, organisation_id: currentMembership.organisation?.id } })
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          .then((res: any) => {
            if (res.success) {
              setCompletedRows((prev) => [index, ...prev]);
              toast.success(
                <ToastContent
                  headline="Invitation sent"
                  text={`Sucessfully invited ${res.user?.email || ''}`}
                />,
              );
            }
            if (res.user?.errors) {
              Object.entries(res.user.errors)
                .filter(([fieldName]) => ['name', 'email', 'role'].includes(fieldName))
                .map(([fieldName, errors]) => {
                  setError(`users.${index}.${fieldName as 'name' | 'email' | 'role'}`, {
                    message: (errors as Array<string>).join(', '),
                  });
                });
            }
            if (res.message?.length > 0 && !res.success) {
              setError(
                `users.${index}.name`,
                { type: 'focus', message: res.message },
                { shouldFocus: true },
              );
            }
          })
      );
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="space-y-4 mb-4">
        {fields.map((field, index) => {
          const completed = completedRows.includes(index);
          return (
            <div key={field.id}>
              <div className="flex space-x-2 w-full">
                {showNameInput && (
                  <Input
                    label="Name"
                    placeholder="Name"
                    wrapperClassName="flex-1"
                    {...register(`users.${index}.name`)}
                    errors={[errors?.users?.[index]?.name?.message]}
                    defaultValue={field.name}
                    disabled={completed}
                  />
                )}
                <Input
                  label="Email"
                  placeholder="Email"
                  wrapperClassName="flex-1"
                  {...register(`users.${index}.email`)}
                  errors={[errors?.users?.[index]?.email?.message]}
                  defaultValue={field.email}
                  disabled={completed}
                />
                <Select
                  label="Role"
                  wrapperClassName=" w-32"
                  {...register(`users.${index}.role`)}
                  errors={[errors?.users?.[index]?.role?.message]}
                  defaultValue={field.role}
                  options={{ user: 'User', admin: 'Admin' }}
                  disabled={completed}
                />
              </div>

              {completed && (
                <span className="text-green-notification">Invitation has been sent.</span>
              )}
            </div>
          );
        })}
      </div>
      <section className="flex justify-between">
        <span
          className="underline font-bold cursor-pointer"
          onClick={() => {
            append({ name: defaultNameValue, email: '', role: 'user' });
          }}>
          + Invite more
        </span>
        <Button type="submit" theme="SUCCESS">
          Invite
        </Button>
      </section>
    </form>
  );
};
export default InviteUserForm;
