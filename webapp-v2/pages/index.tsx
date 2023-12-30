import Spinner from 'components/elements/Spinner/Spinner';
import router from 'next/router';
import { fetchAuthToken } from 'utils/session';

export default function Home(): JSX.Element {
  // @TODO: This should be the landingpage

  if (fetchAuthToken()) {
    router.push('/dashboard');
  } else {
    router.push('/login');
  }

  return <Spinner />;
}
