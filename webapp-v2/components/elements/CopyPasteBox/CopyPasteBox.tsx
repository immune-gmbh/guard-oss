import CopyToClipboardButton from 'components/elements/CopyToClipboardButton/CopyToClipboardButton';

interface ICopyPasteBox extends React.HTMLAttributes<HTMLPreElement> {
  text: string;
}

export default function CopyPasteBox({ text }: ICopyPasteBox): JSX.Element {
  return (
    <pre className="bg-gray-50 p-8 my-4 relative break-all whitespace-pre-wrap">
      <code className="block pr-12">{text}</code>
      <div className="w-8 absolute top-8 right-8 cursor-pointer">
        <CopyToClipboardButton height="8" text={text} />
      </div>
    </pre>
  );
}
