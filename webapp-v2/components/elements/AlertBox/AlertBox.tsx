import Headline from 'components/elements/Headlines/Headline';

interface IAlertBoxProps {
  headline: string;
  text?: string;
}

const AlertBox: React.FC<IAlertBoxProps> = ({ headline, text }) => {
  return (
    <div className="border-4 border-red-cta p-6 text-center space-y-4 bg-cell-small-solid bg-no-repeat bg-right-bottom bg-contain mb-8">
      <Headline size={3}>{headline}</Headline>
      <p>{text}</p>
    </div>
  );
};
export default AlertBox;
