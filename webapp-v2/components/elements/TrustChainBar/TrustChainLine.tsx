import classNames from 'classnames';

interface ITrustChainLine {
  status: boolean;
  unsupported: boolean;
}

export default function TrustChainLine({ status, unsupported }: ITrustChainLine): JSX.Element {
  const classes = classNames(`z-0 border-2 -mx-7 max-h-0 row-start-1`, {
    'border-red-critical border-dashed': !status && !unsupported,
    'border-green-light-notification': status && !unsupported,
    'border-gray-200': unsupported,
  });

  const style = !status
    ? {
        transform: 'matrix(1, 0, 0, 1.5, 0, 0)',
      }
    : null;

  return <div className={classes} style={style}></div>;
}
