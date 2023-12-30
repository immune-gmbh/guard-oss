import Headline from 'components/elements/Headlines/Headline';
import LogEntry, { willBeRendered } from 'components/elements/LogTile/LogEntry';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';
import Spinner from 'components/elements/Spinner/Spinner';
import { useChanges } from 'hooks/changes';
import { useUsers } from 'hooks/users';

interface ILogTileProps {}

const LogTile: React.FC<ILogTileProps> = () => {
  const { data: users, isLoading: loadingUsers } = useUsers();
  const { changes, isLoading: loadingChanges } = useChanges(users || []);

  if (loadingChanges || loadingUsers) {
    return <Spinner />;
  }

  return (
    <SnippetBox>
      <Headline size={4} className="mb-8">
        Latest Activities
      </Headline>
      <div className="space-y-2">
        {changes
          .filter(willBeRendered)
          .slice(0, 3)
          .map((change) => (
            <LogEntry key={change.id} change={change} />
          ))}
        <div className="border-red-cta border-b h-1" />
        {/* TODO: find log design */}
        {/* <Link href="/log" passHref>
          <a>
            <span className="block underline font-bold cursor-pointer">Show log</span>
          </a>
        </Link> */}
      </div>
    </SnippetBox>
  );
};
export default LogTile;
