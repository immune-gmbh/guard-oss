import Headline from 'components/elements/Headlines/Headline';
import SnippetBox from 'components/elements/SnippetBox/SnippetBox';

export default function NoDevices(): JSX.Element {
  return (
    <SnippetBox withBg={true} className="text-center border-red-cta border-4 rounded">
      <Headline as="h3" size={4} bold={true}>
        No devices yet
      </Headline>
      Follow the instructions below to add devices
    </SnippetBox>
  );
}
