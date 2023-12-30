import OnboardingIntro from 'components/elements/OnboardingIntro/OnboardingIntro';
import DashboardLoggedOut from 'components/layouts/dashboard-logged-out';
import { useSession } from 'hooks/useSession';
import { useUpdateUser } from 'hooks/users';
import React, { useEffect } from 'react';

export default function RegistrationWelcome(): JSX.Element {
  const {
    session: {
      user: { id, hasSeenIntro },
    },
  } = useSession();
  const updateUser = useUpdateUser();

  useEffect(() => {
    if (!hasSeenIntro) {
      updateUser.mutate({ id, hasSeenIntro: true });
    }
  }, [hasSeenIntro, id]);

  return (
    <DashboardLoggedOut>
      <OnboardingIntro />
    </DashboardLoggedOut>
  );
}
