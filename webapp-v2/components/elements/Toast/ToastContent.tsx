interface IToastContentProps {
  text: string;
  headline?: string;
}

const ToastContent: React.FC<IToastContentProps> = ({ text, headline }) => {
  return (
    <>
      {headline && <span className="font-bold text-lg">{headline}</span>}
      <span>{text}</span>
    </>
  );
};
export default ToastContent;
