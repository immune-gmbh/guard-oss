/* eslint-disable react-hooks/rules-of-hooks */
import Headline from 'components/elements/Headlines/Headline';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import DashboardLayout from 'components/layouts/dashboard';
import { useSession } from 'hooks/useSession';
import router from 'next/router';

export default function Custom404(): JSX.Element {
  let session;
  try {
    const { session: currentSession } = useSession();
    session = currentSession;
  } catch (error) {
    session = null;
  }

  if (session) {
    return (
      <DashboardLayout>
        <SnippetBox withBg={true} className="text-center border-red-cta border-4 rounded">
          <Headline as="h3" size={4} bold={true}>
            Something went wrong
          </Headline>
          We notified our developers to solve this issue shortly.
        </SnippetBox>
      </DashboardLayout>
    );
  } else {
    router.push('/login');
    return null;
  }
}
