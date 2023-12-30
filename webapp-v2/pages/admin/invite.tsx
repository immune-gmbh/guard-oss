import Actionbar from 'components/elements/Actionbar/Actionbar';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import Select from 'components/elements/Select/Select';
import Spinner from 'components/elements/Spinner/Spinner';
import AdminLayout from 'components/layouts/admin';
import { useOrganisation, useOrganisations } from 'hooks/organisations';
import { useInviteUser } from 'hooks/users';
import { useRouter } from 'next/router';
import { useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { handleLoadError } from 'utils/errorHandling';

function AdminDashboardUsersInvite(): JSX.Element {
  const inviteUser = useInviteUser();
  const router = useRouter();

  const { data: organisations, isLoading, isError } = useOrganisations();
  const { data: organisation } = useOrganisation({ id: router.query.id as string });

  const {
    register,
    handleSubmit,
    control,
    setError,
    formState: { errors },
  } = useForm({
    defaultValues: {
      users: [
        {
          name: '',
          email: '',
          organisation: router.query?.id || '',
          role: organisation?.memberships.length > 0 ? 'user' : 'owner',
        },
      ],
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
          .mutate({ membership: { ...user, organisation_id: user.organisation } })
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          .then((res: any) => {
            if (res.success) {
              setCompletedRows((prev) => [index, ...prev]);
            }
            if (res.user?.errors) {
              Object.entries(res.user.errors)
                .filter(([fieldName]) =>
                  ['name', 'email', 'organisation', 'role'].includes(fieldName),
                )
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

  if (isLoading) return <Spinner />;
  if (isError) return handleLoadError('Organisations');

  const organisationsList = Object.assign(
    {},
    ...organisations.map((org) => ({ [org.id]: org.name })),
  );

  return (
    <AdminLayout>
      <Actionbar>
        <Button onClick={() => router.back()} theme="SECONDARY-RED">
          Go Back
        </Button>
        <Button theme="SUCCESS" onClick={() => handleSubmit(onSubmit)()}>
          Invite User
        </Button>
      </Actionbar>
      <Headline className="mb-6">Invite User</Headline>
      <form role="form" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-4 mb-4">
          {fields.map((field, index) => {
            const completed = completedRows.includes(index);
            return (
              <div key={field.id} data-testid={`row-${index}`}>
                <div className="flex space-x-2 w-full">
                  <Input
                    label="Name"
                    placeholder="Name"
                    wrapperClassName="flex-1"
                    {...register(`users.${index}.name`)}
                    errors={[errors?.users?.[index]?.name?.message]}
                    defaultValue={field.name}
                    disabled={completed}
                  />
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
                    label="Organisation"
                    wrapperClassName="flex-1"
                    {...register(`users.${index}.organisation`)}
                    errors={[errors?.users?.[index]?.organisation?.message]}
                    defaultValue={field.organisation}
                    options={organisationsList}
                    disabled={completed}
                  />
                  <Select
                    label="Role"
                    wrapperClassName="flex-1"
                    {...register(`users.${index}.role`)}
                    errors={[errors?.users?.[index]?.role?.message]}
                    defaultValue={field.role}
                    options={
                      organisation?.memberships.length > 0
                        ? { owner: 'Owner', user: 'User', admin: 'Admin' }
                        : { owner: 'Owner' }
                    }
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
        <section>
          <span
            className="underline font-bold cursor-pointer"
            onClick={() => {
              append({ name: '', email: '', role: 'user' });
            }}>
            + Add user
          </span>
        </section>
      </form>
    </AdminLayout>
  );
}
export default AdminDashboardUsersInvite;
