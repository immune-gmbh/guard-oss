import Actionbar from 'components/elements/Actionbar/Actionbar';
import { EuropeanCountries, Countries } from 'components/elements/BillingInfoForm/Countries';
import Button from 'components/elements/Button/Button';
import Headline from 'components/elements/Headlines/Headline';
import Input from 'components/elements/Input/Input';
import LoadingSpinner from 'components/elements/LoadingSpinner/LoadingSpinner';
import Select from 'components/elements/Select/Select';
import Spinner from 'components/elements/Spinner/Spinner';
import ControlledToggle from 'components/elements/Toggle/ControlledToggle';
import NextJsRoutes from 'generated/NextJsRoutes';
import { JsRoutesRails } from 'generated/authsrvRoutes';
import { useCreateOrganisation, useOrganisation, useUpdateOrganisation } from 'hooks/organisations';
import { useSession } from 'hooks/useSession';
import { camel } from 'kitsu-core';
import { useRouter } from 'next/router';
import { SelectedDevicesContext } from 'provider/SelectedDevicesProvider';
import React, { useContext, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { DeviceActionType } from 'reducer/SelectedDevicesReducer';
import { useSWRConfig } from 'swr';
import { refreshSession } from 'utils/api';
import { SerializedMembership, SerializedOrganisation, Response, Err } from 'utils/types';

const editableOrganisationAttributes = ['freeloader', 'name', 'invoiceName', 'vatNumber'] as const;

interface IEditOrganisationInfo {
  create?: boolean;
  admin?: boolean;
  afterSubmit?: (info: any) => void;
}

function EditOrganisationInfo({ create, admin, afterSubmit }: IEditOrganisationInfo): JSX.Element {
  const router = useRouter();
  const { setSession, session, setCurrentMembership } = useSession();
  const updateOrganisation = useUpdateOrganisation();
  const createOrganisation = useCreateOrganisation();
  const selectedDevices = useContext(SelectedDevicesContext);
  const { mutate } = useSWRConfig();
  const { data: organisation } = useOrganisation({ id: router.query.id as string });
  const {
    register,
    formState: { errors, isDirty },
    setValue,
    setError,
    getValues,
    control,
    handleSubmit,
    reset,
  } = useForm();

  const [isLoading, setIsLoading] = useState(false);

  // reload the form when default values change
  useEffect(() => {
    // if address values are undefined, the form keeps the former organisation's values
    reset(
      {
        ...organisation,
        address: {
          streetAndNumber: organisation?.address?.streetAndNumber || null,
          postalCode: organisation?.address?.postalCode || null,
          city: organisation?.address?.city || null,
          country: organisation?.address?.country || null,
        },
      },
      { keepIsSubmitted: true, keepSubmitCount: true },
    );
  }, [organisation, reset]);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onSubmit = async (data: Record<string, any>): Promise<any> => {
    setIsLoading(true);

    const submitFunction = create ? createOrganisation : updateOrganisation;

    data.vatNumber = EuropeanCountries.includes(data?.address?.country) ? data?.vatNumber : '';
    submitFunction
      .mutate({ id: router.query.id, ...data })
      .then(async (organisation: Response<SerializedOrganisation>) => {
        const { errors } = organisation;
        let hasErrors = false;

        errors &&
          errors.map((err: Err) => {
            hasErrors = true;
            let name = '';
            if (err.id) {
              name = camel(err.id) as typeof editableOrganisationAttributes[number];
            }
            setError(name, {
              message: err.title,
            });
          });

        if (!hasErrors) {
          if (create) {
            toast.success('Successfully created new organisation');
            if (admin) {
              router.push({
                pathname: NextJsRoutes.adminOrganisationsIdUsersPath,
                query: { id: organisation.id },
              });
            } else {
              const usableMemberships = [];
              for (const m of organisation?.memberships || []) {
                // can this membership be used for authentication?
                if (!m.token || !m.enrollmentToken) {
                  continue;
                }

                // recreate recursive relationship between organisation and
                // membership
                if (!m.organisation) {
                  m.organisation = Object.assign({}, organisation);
                  m.organisation.memberships = [];
                }

                usableMemberships.push(m);

                // merge new membership into session
                let newMembership = true;
                for (const mm of session?.memberships || []) {
                  newMembership &&= mm.id !== m.id;
                }
                if (newMembership) {
                  session.memberships.push(m as SerializedMembership);
                }
              }

              if (usableMemberships.length === 1) {
                setSession(session);
                selectedDevices.dispatch({ type: DeviceActionType.CLEAR });
                setCurrentMembership(usableMemberships[0]);
                mutate(JsRoutesRails.v2_membership_path({ id: usableMemberships[0].id }));
              }
            }
          } /* !create */ else {
            mutate(
              JsRoutesRails.v2_organisation_path({ id: organisation.id }),
              organisation,
              false,
            );
            await refreshSession(setSession, {
              ...session.currentMembership,
              organisation,
            });
            toast.success('Organisation saved.');
          }
        } else {
          toast.error('Something went wrong while updating your organisation');
        }

        if (afterSubmit) {
          afterSubmit({ create, errors, organisation });
        }
        setIsLoading(false);
      });
  };

  register('address.country', {
    required: { value: true, message: 'Field is required' },
  });

  const [showVatIdField, setShowVatIdField] = useState(
    EuropeanCountries.includes(organisation?.address?.country),
  );
  const countryOptions = Object.assign({}, ...Countries.map((c) => ({ [c]: c })));

  return (
    <>
      {isLoading && <Spinner />}
      <form
        id="EditOrganisationInfo"
        role="form"
        onSubmit={handleSubmit(onSubmit)}
        className="space-y-4">
        <Headline size={4}>Regular Information</Headline>
        <Input
          label="Organisation Name"
          placeholder="Organisation Name"
          {...register('name')}
          errors={[errors?.name?.message]}
        />
        {create && admin && (
          <ControlledToggle control={control as any} label="Freeloader" name="freeloader" />
        )}
        <hr />
        <Headline size={4}>Organisation Contact</Headline>
        <div className="space-y-2 flex flex-col">
          <Input
            label="Name for invoice"
            placeholder="Name"
            {...register('invoiceName')}
            errors={[errors?.invoiceName?.message]}
          />
          <Input
            label="Street & Number"
            placeholder="Street & Number"
            {...register('address.streetAndNumber', {
              required: { value: true, message: 'Field is required' },
            })}
            errors={[errors?.address?.streetAndNumber?.message]}
          />
          <Input
            label="Postal Code"
            placeholder="Postal Code"
            {...register('address.postalCode', {
              required: { value: true, message: 'Field is required' },
            })}
            errors={[errors?.address?.postalCode?.message]}
          />
          <Input
            label="City"
            placeholder="City"
            {...register('address.city', {
              required: { value: true, message: 'Field is required' },
            })}
            errors={[errors?.address?.city?.message]}
          />
          <Select
            options={countryOptions}
            qaLabel="country"
            label="Country"
            placeholder="Country"
            selectedOption={organisation?.address?.country}
            onChange={(e) => {
              setValue(
                'address.country',
                (e.target as HTMLSelectElement).selectedOptions[0].value,
                {
                  shouldValidate: true,
                },
              );
              setShowVatIdField(EuropeanCountries.includes(getValues('address.country')));
            }}
            errors={[errors?.address?.country?.message]}
          />
          {showVatIdField && (
            <Input
              label="EU VAT-ID"
              placeholder="EU VAT-ID"
              {...register('vatNumber')}
              errors={[errors?.vatNumber?.message]}
            />
          )}
          {create && admin && <Input type="hidden" value="true" {...register('admin_view')} />}
        </div>
        <Actionbar>
          <div />
          {isDirty && (
            <Button
              theme="SUCCESS"
              type="submit"
              form="EditOrganisationInfo"
              disabled={isLoading}
              data-qa="save-orga">
              {isLoading && <LoadingSpinner />}
              {create ? 'Create' : 'Save'} Organisation
            </Button>
          )}
        </Actionbar>
      </form>
    </>
  );
}
export default EditOrganisationInfo;
